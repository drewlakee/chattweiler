package vk

import (
	"chattweiler/internal/logging"
	"chattweiler/internal/repository"
	"chattweiler/internal/repository/model"
	"strconv"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
)

type Checker struct {
	// https://dev.vk.com/method/messages.getConversationsById
	// conversationId = 2000000000 + id, id - chat id
	conversationId         int64
	communityId            int64
	checkInterval          time.Duration
	gracePeriod            time.Duration
	vkapi                  *api.VK
	phrasesRepo            repository.PhraseRepository
	membershipWarningsRepo repository.MembershipWarningRepository
}

func NewChecker(
	conversationId,
	communityId int64,
	checkInterval,
	gracePeriod time.Duration,
	vkapi *api.VK,
	phrasesRepo repository.PhraseRepository,
	membershipWarningsRepo repository.MembershipWarningRepository,
) *Checker {
	return &Checker{
		conversationId:         conversationId,
		communityId:            communityId,
		checkInterval:          checkInterval,
		gracePeriod:            gracePeriod,
		vkapi:                  vkapi,
		phrasesRepo:            phrasesRepo,
		membershipWarningsRepo: membershipWarningsRepo,
	}
}

func (checker *Checker) checkAlreadyRelevantMembershipWarnings(members map[int]object.UsersUser) (map[int]bool, error) {
	alreadyForewarnedUsers := map[int]bool{}
	relevantWarnings := checker.membershipWarningsRepo.FindAllRelevant()

	var expiredWarnings []model.MembershipWarning
	for _, warning := range relevantWarnings {
		gracePeriod, _ := time.ParseDuration(warning.GracePeriod)
		if time.Now().After(warning.FirstWarningTs.Add(gracePeriod)) {
			expiredWarnings = append(expiredWarnings, warning)
		}
		alreadyForewarnedUsers[warning.UserID] = true
	}

	var membershipVector api.GroupsIsMemberUserIDsResponse
	if len(expiredWarnings) > 0 {
		usersWithWarning := make([]int, len(expiredWarnings))
		isMemberUserIDsBuilder := params.NewGroupsIsMemberBuilder()
		isMemberUserIDsBuilder.GroupID(strconv.FormatInt(checker.communityId, 10))
		for _, userWithWarning := range expiredWarnings {
			usersWithWarning = append(usersWithWarning, userWithWarning.UserID)
		}
		isMemberUserIDsBuilder.UserIDs(usersWithWarning)
		membershipVectorResponse, err := checker.vkapi.GroupsIsMemberUserIDs(isMemberUserIDsBuilder.Params)
		membershipVector = membershipVectorResponse
		if err != nil {
			return nil, err
		}
	}

	for index, expiredWarning := range expiredWarnings {
		if membershipVector != nil {
			_, stillSittingInChat := members[membershipVector[index].UserID]
			if stillSittingInChat && !bool(membershipVector[index].Member) {
				messagesRemoveChatUserBuilder := params.NewMessagesRemoveChatUserBuilder()
				messagesRemoveChatUserBuilder.UserID(expiredWarning.UserID)
				messagesRemoveChatUserBuilder.ChatID(int(checker.conversationId))
				_, err := checker.vkapi.MessagesRemoveChatUser(messagesRemoveChatUserBuilder.Params)
				if err != nil && err.Error() != "api: User not found in chat" {
					return nil, err
				}
			}
		}
	}

	if len(expiredWarnings) > 0 {
		checker.membershipWarningsRepo.UpdateAllToIrrelevant(expiredWarnings...)
	}

	return alreadyForewarnedUsers, nil
}

func (checker *Checker) checkChatForNewWarning(members map[int]object.UsersUser, alreadyForewarnedUsers map[int]bool) error {
	isMemberUserIDsBuilder := params.NewGroupsIsMemberBuilder()
	isMemberUserIDsBuilder.GroupID(strconv.FormatInt(checker.communityId, 10))
	userIds := make([]int, len(members))
	index := 0
	for userId := range members {
		userIds[index] = userId
		index++
	}
	if len(members) == 0 {
		return nil
	}

	isMemberUserIDsBuilder.UserIDs(userIds)
	membershipVector, err := checker.vkapi.GroupsIsMemberUserIDs(isMemberUserIDsBuilder.Params)
	if err != nil {
		return err
	}

	for _, membership := range membershipVector {
		_, alreadyForewarnedUser := alreadyForewarnedUsers[membership.UserID]
		if !bool(membership.Member) && !alreadyForewarnedUser {
			userProfile := members[membership.UserID]

			newWarning := model.MembershipWarning{}
			newWarning.IsRelevant = true
			newWarning.GracePeriod = checker.gracePeriod.String()
			newWarning.FirstWarningTs = time.Now()
			newWarning.Username = userProfile.ScreenName
			newWarning.UserID = userProfile.ID
			checker.membershipWarningsRepo.Insert(newWarning)

			phrases := checker.phrasesRepo.FindAllByType(model.MembershipWarningType)
			if len(phrases) == 0 {
				logging.Log.Warn(logPackage, "Checker.checkChatForNewWarning", "there's no membership warning phrases, message won't be sent")
				return nil
			}

			peerId := 2000000000 + int(checker.conversationId)
			messageToSend := BuildMessageUsingPersonalizedPhrase(peerId, &userProfile, phrases)
			_, err := checker.vkapi.MessagesSend(messageToSend)
			if err != nil {
				logging.Log.Error(logPackage, "Checker.checkChatForNewWarning", err, "message sending error. Sent params: %v", messageToSend)
				return err
			}

			return nil
		}
	}

	return nil
}

func (checker *Checker) LoopCheck() {
	successfulCheckAttempt := true

	for {
		// made in order to leverage cpu load
		// if constant error happens
		if !successfulCheckAttempt {
			time.Sleep(checker.checkInterval)
		}

		conversationMembers, err := checker.vkapi.MessagesGetConversationMembers(api.Params{
			"peer_id": 2000000000 + checker.conversationId,
		})
		if err != nil {
			logging.Log.Error(logPackage, "Checker.LoopCheck", err, "vk api error")
			successfulCheckAttempt = false
			continue
		}

		members := filterOnlyCommonMembers(conversationMembers)
		alreadyForewarnedUsers, err := checker.checkAlreadyRelevantMembershipWarnings(members)
		if err != nil {
			logging.Log.Error(logPackage, "Checker.LoopCheck", err, "error occurred during relevant membership warnings fetching")
			successfulCheckAttempt = false
			continue
		}

		err = checker.checkChatForNewWarning(members, alreadyForewarnedUsers)
		if err != nil {
			logging.Log.Error(logPackage, "Checker.LoopCheck", err, "error occurred during new membership warnings checking")
			successfulCheckAttempt = false
			continue
		}

		successfulCheckAttempt = true
		time.Sleep(checker.checkInterval)
	}
}

func filterOnlyCommonMembers(response api.MessagesGetConversationMembersResponse) map[int]object.UsersUser {
	commonMembers := make(map[int]bool)
	for _, member := range response.Items {
		if !member.IsAdmin && !member.IsOwner && member.CanKick {
			commonMembers[member.MemberID] = true
		}
	}

	commonMemberUserProfiles := make(map[int]object.UsersUser, len(commonMembers))
	for _, user := range response.Profiles {
		if _, isCommonMember := commonMembers[user.ID]; isCommonMember {
			commonMemberUserProfiles[user.ID] = user
		}
	}

	return commonMemberUserProfiles
}
