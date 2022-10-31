package service

import (
	"chattweiler/internal/logging"
	"chattweiler/internal/repository"
	"chattweiler/internal/utils"
	"chattweiler/internal/vk"
	"chattweiler/internal/vk/content"
	"math/rand"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
)

type CachedRandomAttachmentsContentCollector struct {
	client               *api.VK
	contentCommandId     int
	contentSourceRepo    repository.ContentCommandRepository
	cachedAttachments    map[vk.MediaAttachmentType][]content.MediaAttachment
	maxContentFetchBound int
	attachmentTypes      []vk.MediaAttachmentType
}

func NewCachedRandomAttachmentsContentCollector(
	client *api.VK,
	attachmentTypes []vk.MediaAttachmentType,
	contentCommandId int,
	contentSourceRepo repository.ContentCommandRepository,
) *CachedRandomAttachmentsContentCollector {
	return &CachedRandomAttachmentsContentCollector{
		client:            client,
		contentCommandId:  contentCommandId,
		contentSourceRepo: contentSourceRepo,
		cachedAttachments: make(map[vk.MediaAttachmentType][]content.MediaAttachment),
		attachmentTypes:   attachmentTypes,

		// https://dev.vk.com/method/wall.get#count parameters' constraints
		maxContentFetchBound: 100,
	}
}

func (collector *CachedRandomAttachmentsContentCollector) CollectOne() *content.MediaAttachment {
	rand.Seed(time.Now().UnixNano())
	attachmentType := collector.attachmentTypes[rand.Intn(len(collector.attachmentTypes))]
	threshold := int(float32(getMaxCachedAttachments(attachmentType)) * getCacheRefreshThresholdFor(attachmentType))
	attachments, exists := collector.cachedAttachments[attachmentType]
	if !exists || len(attachments) <= threshold {
		collector.refreshCacheDifference(attachmentType)
		if len(collector.cachedAttachments[attachmentType]) == 0 {
			logging.Log.Warn(logPackage, "CachedRandomAttachmentsContentCollector.CollectOne", "empty attachments. attachmentsType=%s, contentCommandId=%d", attachmentType, collector.contentCommandId)
			return nil
		}
	}

	return collector.getAndRemoveCachedAttachment(attachmentType)
}

func (collector *CachedRandomAttachmentsContentCollector) getAndRemoveCachedAttachment(attachmentType vk.MediaAttachmentType) *content.MediaAttachment {
	attachments := collector.cachedAttachments[attachmentType]
	randomCachedAttachmentIndex := rand.Intn(len(attachments))
	attachment := attachments[randomCachedAttachmentIndex]

	// swap last with random chosen one
	lastAttachment := attachments[len(attachments)-1]
	attachments[len(attachments)-1] = attachments[randomCachedAttachmentIndex]
	attachments[randomCachedAttachmentIndex] = lastAttachment

	// and cut off the tail of slice
	collector.cachedAttachments[attachmentType] = attachments[:len(attachments)-1]
	return &attachment
}

func (collector *CachedRandomAttachmentsContentCollector) refreshCacheDifference(attachmentType vk.MediaAttachmentType) {
	contentCommand := collector.contentSourceRepo.FindById(collector.contentCommandId)
	randomVkCommunity := collector.getCommunity(contentCommand.ContentDescriptor.CommunitySourceIDs)

	count, err := vk.GetWallPostsCount(collector.client, randomVkCommunity)
	if err != nil {
		logging.Log.Error(logPackage, "CachedRandomAttachmentsContentCollector.refreshCacheDifference", err, "vk api error")
	}

	randomSequenceFetchOffset := collector.getRandomWallPostsOffset(count, collector.maxContentFetchBound)
	contentSequence := collector.fetchContentSequence(attachmentType, randomVkCommunity, randomSequenceFetchOffset, collector.maxContentFetchBound)
	collector.cachedAttachments[attachmentType] = append(collector.cachedAttachments[attachmentType], collector.gatherDifference(contentSequence, attachmentType)...)
}

func (collector *CachedRandomAttachmentsContentCollector) gatherDifference(
	contentSequence []object.WallWallpostAttachment,
	attachmentType vk.MediaAttachmentType,
) []content.MediaAttachment {
	alreadyPickedUpContentVector := make([]int, len(contentSequence))
	alreadyPickedUpContentCount := 0

	difference := getMaxCachedAttachments(attachmentType) - len(collector.cachedAttachments[attachmentType])
	var collectResult []content.MediaAttachment
	for alreadyPickedUpContentCount != len(contentSequence) && difference > 0 {
		randomIndex := rand.Intn(len(contentSequence))
		for alreadyPickedUpContentVector[randomIndex] == 1 {
			randomIndex++
			if randomIndex == len(alreadyPickedUpContentVector) {
				randomIndex = 0
			}
		}

		collectResult = append(collectResult, content.MediaAttachment{
			Type: attachmentType,
			Data: &contentSequence[randomIndex],
		})
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
	return utils.Clamp[int](randomSequenceFetchOffset, 0, wallPostsCount)
}

func (collector *CachedRandomAttachmentsContentCollector) fetchContentSequence(
	attachmentType vk.MediaAttachmentType,
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
		logging.Log.Error(logPackage, "CachedRandomAttachmentsContentCollector.fetchContentSequence", err, "empty content sequence")
		return []object.WallWallpostAttachment{}
	}

	var attachments []object.WallWallpostAttachment
	for _, wallPost := range response.Items {
		for _, attachment := range wallPost.Attachments {
			if attachment.Type == string(attachmentType) &&
				isSharingEnabled(attachmentType, attachment) &&
				len(attachments) < count {
				attachments = append(attachments, attachment)
				break
			}
		}
	}

	return attachments
}

func isSharingEnabled(attachmentsType vk.MediaAttachmentType, attachment object.WallWallpostAttachment) bool {
	switch attachmentsType {
	case vk.VideoType:
		return bool(attachment.Video.CanRepost)
	}
	return true
}

func (collector *CachedRandomAttachmentsContentCollector) getCommunity(communities []string) string {
	rand.Seed(time.Now().UnixNano())
	return communities[rand.Intn(len(communities))]
}
