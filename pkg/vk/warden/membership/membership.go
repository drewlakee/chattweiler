package membership

import (
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/repository/model/types"
	"chattweiler/pkg/vk/messages"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"strconv"
	"time"
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

func (checker *Checker) checkAlreadyRelevantMembershipWarnings(conversationMembers api.MessagesGetConversationMembersResponse) (map[int]bool, error) {
	alreadyForewarnedUsers := map[int]bool{}
	relevantWarnings := checker.membershipWarningsRepo.FindAllRelevant()

	var expiredWarnings []model.MembershipWarning
	for _, warning := range relevantWarnings {
		if time.Now().After(warning.FirstWarningTs.Add(warning.GracePeriod)) {
			expiredWarnings = append(expiredWarnings, warning)
		} else {
			alreadyForewarnedUsers[warning.UserID] = true
		}
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
		if membershipVector == nil || !membershipVector[index].Member {
			messagesRemoveChatUserBuilder := params.NewMessagesRemoveChatUserBuilder()
			messagesRemoveChatUserBuilder.UserID(expiredWarning.UserID)
			messagesRemoveChatUserBuilder.ChatID(int(checker.conversationId))
			_, err := checker.vkapi.MessagesRemoveChatUser(messagesRemoveChatUserBuilder.Params)
			if err != nil && err.Error() != "api: User not found in chat" {
				return nil, err
			}
		}
	}

	if len(expiredWarnings) > 0 {
		checker.membershipWarningsRepo.UpdateAllToUnRelevant(expiredWarnings...)
	}

	return alreadyForewarnedUsers, nil
}

func (checker *Checker) checkChatForNewWarning(convesationMembers api.MessagesGetConversationMembersResponse, alreadyForewarnedUsers map[int]bool) error {
	isMemberUserIDsBuilder := params.NewGroupsIsMemberBuilder()
	isMemberUserIDsBuilder.GroupID(strconv.FormatInt(checker.communityId, 10))
	userIds := make([]int, convesationMembers.Count-1)
	for index, user := range convesationMembers.Items {
		// community id is negative
		// except the community
		if user.MemberID > 0 {
			userIds[index-1] = user.MemberID
		}
	}
	if len(userIds) == 0 {
		fmt.Println("No one in the conversation except the community")
		return nil
	}

	isMemberUserIDsBuilder.UserIDs(userIds)
	membershipVector, err := checker.vkapi.GroupsIsMemberUserIDs(isMemberUserIDsBuilder.Params)
	if err != nil {
		return err
	}

	for index, membership := range membershipVector {
		_, alreadyForewarnedUser := alreadyForewarnedUsers[membership.UserID]
		if !bool(convesationMembers.Items[index+1].IsAdmin) && !bool(membership.Member) && !alreadyForewarnedUser {
			newWarning := model.MembershipWarning{}
			newWarning.IsRelevant = true
			newWarning.GracePeriod = checker.gracePeriod
			newWarning.FirstWarningTs = time.Now()
			newWarning.Username = convesationMembers.Profiles[index].ScreenName
			newWarning.UserID = membership.UserID
			checker.membershipWarningsRepo.Insert(newWarning)

			peerId := 2000000000 + int(checker.conversationId)
			_, err := checker.vkapi.MessagesSend(messages.BuildMessageUsingPersonalizedPhrase(
				peerId,
				&convesationMembers.Profiles[index],
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

		convesationMembers, err := checker.vkapi.MessagesGetConversationMembers(api.Params{
			"peer_id": 2000000000 + checker.conversationId,
		})
		if err != nil {
			fmt.Println(err)
			successfulCheckAttempt = false
			continue
		}

		alreadyForewarnedUsers, err := checker.checkAlreadyRelevantMembershipWarnings(convesationMembers)
		if err != nil {
			fmt.Println(err)
			successfulCheckAttempt = false
			continue
		}

		err = checker.checkChatForNewWarning(convesationMembers, alreadyForewarnedUsers)
		if err != nil {
			fmt.Println(err)
			successfulCheckAttempt = false
			continue
		}

		successfulCheckAttempt = true
		time.Sleep(checker.checkInterval)
	}
}
