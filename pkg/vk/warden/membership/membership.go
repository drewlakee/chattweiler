package membership

import (
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/repository/model/types"
	"chattweiler/pkg/vk/messages"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var packageLogFields = logrus.Fields{
	"package": "membership",
}

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
		if time.Now().After(warning.FirstWarningTs.Add(warning.GracePeriod)) {
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
		checker.membershipWarningsRepo.UpdateAllToUnRelevant(expiredWarnings...)
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
			newWarning.GracePeriod = checker.gracePeriod
			newWarning.FirstWarningTs = time.Now()
			newWarning.Username = userProfile.ScreenName
			newWarning.UserID = userProfile.ID
			checker.membershipWarningsRepo.Insert(newWarning)

			peerId := 2000000000 + int(checker.conversationId)
			_, err := checker.vkapi.MessagesSend(messages.BuildMessageUsingPersonalizedPhrase(
				peerId,
				&userProfile,
				types.MembershipWarning,
				checker.phrasesRepo.FindAllByType(types.MembershipWarning),
			))
			if err != nil {
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
		if !successfulCheckAttempt {
			time.Sleep(checker.checkInterval)
		}

		conversationMembers, err := checker.vkapi.MessagesGetConversationMembers(api.Params{
			"peer_id": 2000000000 + checker.conversationId,
		})
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func": "LoopCheck",
				"err":  err,
			}).Error()
			successfulCheckAttempt = false
			continue
		}

		members := filterOnlyCommonMembers(conversationMembers)
		alreadyForewarnedUsers, err := checker.checkAlreadyRelevantMembershipWarnings(members)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func": "LoopCheck",
				"err":  err,
			}).Error()
			successfulCheckAttempt = false
			continue
		}

		err = checker.checkChatForNewWarning(members, alreadyForewarnedUsers)
		if err != nil {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"func": "LoopCheck",
				"err":  err,
			}).Error()
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
