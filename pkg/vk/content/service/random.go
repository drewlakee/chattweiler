package service

import (
	"chattweiler/pkg/repository"
	"chattweiler/pkg/utils"
	"chattweiler/pkg/vk"
	"math/rand"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
)

type CachedRandomAttachmentsContentCollector struct {
	client                *api.VK
	attachmentsType       vk.AttachmentsType
	contentCommand        string
	contentSourceRepo     repository.ContentCommandRepository
	maxCachedAttachments  int
	cachedAttachments     []object.WallWallpostAttachment
	cacheRefreshThreshold float32
	maxContentFetchBound  int
}

func NewCachedRandomAttachmentsContentCollector(
	client *api.VK,
	attachmentsType vk.AttachmentsType,
	contentCommand string,
	contentSourceRepo repository.ContentCommandRepository,
	maxCachedAttachments int,
	cacheRefreshThreshold float32,
) *CachedRandomAttachmentsContentCollector {
	return &CachedRandomAttachmentsContentCollector{
		client:                client,
		attachmentsType:       attachmentsType,
		contentCommand:        contentCommand,
		contentSourceRepo:     contentSourceRepo,
		maxCachedAttachments:  maxCachedAttachments,
		cachedAttachments:     nil,
		cacheRefreshThreshold: utils.ClampFloat32(cacheRefreshThreshold, 0.0, 1.0),

		// https://dev.vk.com/method/wall.get#count parameters' constraints
		maxContentFetchBound: 100,
	}
}

func (collector *CachedRandomAttachmentsContentCollector) CollectOne() object.WallWallpostAttachment {
	rand.Seed(time.Now().UnixNano())
	threshold := int(float32(collector.maxCachedAttachments) * collector.cacheRefreshThreshold)
	if len(collector.cachedAttachments) <= threshold {
		collector.refreshCacheDifference()
		if len(collector.cachedAttachments) == 0 {
			logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
				"struct":   "CachedRandomAttachmentsContentCollector",
				"func":     "CollectOne",
				"fallback": "empty wall post attachment",
			}).Warn()
			return object.WallWallpostAttachment{}
		}
	}

	return collector.getAndRemoveCachedAttachment()
}

func (collector *CachedRandomAttachmentsContentCollector) getAndRemoveCachedAttachment() object.WallWallpostAttachment {
	randomCachedAttachmentIndex := rand.Intn(len(collector.cachedAttachments))
	attachment := collector.cachedAttachments[randomCachedAttachmentIndex]

	// swap last with random chosen one
	lastAttachment := collector.cachedAttachments[len(collector.cachedAttachments)-1]
	collector.cachedAttachments[len(collector.cachedAttachments)-1] = collector.cachedAttachments[randomCachedAttachmentIndex]
	collector.cachedAttachments[randomCachedAttachmentIndex] = lastAttachment

	// and cut off the tail of slice
	collector.cachedAttachments = collector.cachedAttachments[:len(collector.cachedAttachments)-1]
	return attachment
}

func (collector *CachedRandomAttachmentsContentCollector) refreshCacheDifference() {
	contentCommand := collector.contentSourceRepo.FindByCommand(collector.contentCommand)
	randomVkCommunity := collector.getRandomVkCommunity(contentCommand.GetSeparatedCommunityIDs())
	wallPostsCount := collector.getWallPostsCount(randomVkCommunity)
	randomSequenceFetchOffset := collector.getRandomWallPostsOffset(wallPostsCount, collector.maxContentFetchBound)
	contentSequence := collector.fetchContentSequence(randomVkCommunity, randomSequenceFetchOffset, collector.maxContentFetchBound)
	collector.cachedAttachments = append(collector.cachedAttachments, collector.gatherDifference(contentSequence)...)
}

func (collector *CachedRandomAttachmentsContentCollector) gatherDifference(
	contentSequence []object.WallWallpostAttachment,
) []object.WallWallpostAttachment {
	alreadyPickedUpContentVector := make([]int, len(contentSequence))
	alreadyPickedUpContentCount := 0

	difference := collector.maxCachedAttachments - len(collector.cachedAttachments)
	var collectResult []object.WallWallpostAttachment
	for alreadyPickedUpContentCount != len(contentSequence) && difference > 0 {
		randomIndex := rand.Intn(len(contentSequence))
		for alreadyPickedUpContentVector[randomIndex] == 1 {
			randomIndex++
			if randomIndex == len(alreadyPickedUpContentVector) {
				randomIndex = 0
			}
		}

		collectResult = append(collectResult, contentSequence[randomIndex])
		alreadyPickedUpContentVector[randomIndex] = 1

		alreadyPickedUpContentCount++
		difference--
	}
	return collectResult
}

func (collector *CachedRandomAttachmentsContentCollector) getRandomWallPostsOffset(wallPostsCount, maxContentFetchBound int) int {
	rand.Seed(time.Now().UnixNano())
	randomSequenceFetchOffset := rand.Intn(wallPostsCount)
	if (wallPostsCount - randomSequenceFetchOffset) < maxContentFetchBound {
		randomSequenceFetchOffset -= maxContentFetchBound - (wallPostsCount - randomSequenceFetchOffset)
	}
	if randomSequenceFetchOffset < 0 {
		randomSequenceFetchOffset = 1
	}
	return randomSequenceFetchOffset
}

func (collector *CachedRandomAttachmentsContentCollector) fetchContentSequence(
	community string,
	offset,
	count int,
) []object.WallWallpostAttachment {
	response, err := collector.client.WallGet(api.Params{
		"domain": community,
		"count":  count,
		"offset": offset,
	})

	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CachedRandomAttachmentsContentCollector",
			"func":     "fetchContentSequence",
			"err":      err,
			"fallback": "empty content sequence",
		}).Error()
		return []object.WallWallpostAttachment{}
	}

	var attachments []object.WallWallpostAttachment
	for _, wallPost := range response.Items {
		for _, attachment := range wallPost.Attachments {
			if attachment.Type == string(collector.attachmentsType) &&
				canShareInChat(collector.attachmentsType, attachment) &&
				len(attachments) < count {
				attachments = append(attachments, attachment)
				break
			}
		}
	}

	return attachments
}

func canShareInChat(attachmentsType vk.AttachmentsType, attachment object.WallWallpostAttachment) bool {
	switch attachmentsType {
	case vk.VideoType:
		return bool(attachment.Video.CanRepost)
	}
	return true
}

func (collector *CachedRandomAttachmentsContentCollector) getRandomVkCommunity(communities []string) string {
	rand.Seed(time.Now().UnixNano())
	return communities[rand.Intn(len(communities))]
}

func (collector *CachedRandomAttachmentsContentCollector) getWallPostsCount(community string) int {
	response, err := collector.client.WallGet(api.Params{
		"domain": community,
		"count":  1,
	})

	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CachedRandomAttachmentsContentCollector",
			"func":     "getWallPostsCount",
			"err":      err,
			"fallback": "0",
		}).Error()
	}

	return response.Count
}
