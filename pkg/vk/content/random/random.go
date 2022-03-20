package random

import (
	"chattweiler/pkg/repository"
	"chattweiler/pkg/repository/model"
	"chattweiler/pkg/repository/model/types"
	"chattweiler/pkg/utils/math"
	"chattweiler/pkg/vk/content"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

var packageLogFields = logrus.Fields{
	"package": "random",
}

type CachedRandomAttachmentsContentCollector struct {
	client                *api.VK
	attachmentsType       content.AttachmentsType
	contentSourceRepo     repository.ContentSourceRepository
	maxCachedAttachments  int
	cachedAttachments     []object.WallWallpostAttachment
	cacheRefreshThreshold float32
	maxContentFetchBound  int
}

func NewCachedRandomAttachmentsContentCollector(
	client *api.VK,
	attachmentsType content.AttachmentsType,
	contentSourceRepo repository.ContentSourceRepository,
	maxCachedAttachments int,
	cacheRefreshThreshold float32,
) *CachedRandomAttachmentsContentCollector {
	return &CachedRandomAttachmentsContentCollector{
		client:                client,
		attachmentsType:       attachmentsType,
		contentSourceRepo:     contentSourceRepo,
		maxCachedAttachments:  maxCachedAttachments,
		cachedAttachments:     nil,
		cacheRefreshThreshold: math.ClampFloat32(cacheRefreshThreshold, 0.0, 1.0),

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
	source := collector.getRandomSource()
	wallPostsCount := collector.getSourceWallPostsCount(source)
	randomSequenceFetchOffset := collector.getRandomWallPostsOffset(wallPostsCount, collector.maxContentFetchBound)
	contentSequence := collector.fetchContentSequence(source, randomSequenceFetchOffset, collector.maxContentFetchBound)
	alreadyPickedUpContentVector := make([]int, len(contentSequence))
	alreadyPickedUpContentCount := 0

	difference := collector.maxCachedAttachments - len(collector.cachedAttachments)
	var collectResult []object.WallWallpostAttachment
	for difference > 0 || alreadyPickedUpContentCount != len(contentSequence) {
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

	for _, attachment := range collectResult {
		collector.cachedAttachments = append(collector.cachedAttachments, attachment)
	}
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

func (collector *CachedRandomAttachmentsContentCollector) fetchContentSequence(source model.ContentSource, offset, count int) []object.WallWallpostAttachment {
	response, err := collector.client.WallGet(api.Params{
		"domain": source.VkCommunityID,
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
			if attachment.Type == string(collector.attachmentsType) && len(attachments) < count {
				attachments = append(attachments, attachment)
			}
		}
	}

	return attachments
}

func (collector *CachedRandomAttachmentsContentCollector) getRandomSource() model.ContentSource {
	rand.Seed(time.Now().UnixNano())

	var contentSources []model.ContentSource
	switch collector.attachmentsType {
	case content.Audio:
		contentSources = collector.contentSourceRepo.FindAllByType(types.Audio)
	case content.Photo:
		contentSources = collector.contentSourceRepo.FindAllByType(types.Picture)
	default:
		contentSources = []model.ContentSource{}
	}

	if len(contentSources) == 0 {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CachedRandomAttachmentsContentCollector",
			"func":     "getRandomSource",
			"fallback": "empty content source",
		}).Warn()
		return model.ContentSource{}
	} else {
		return contentSources[rand.Intn(len(contentSources))]
	}
}

func (collector *CachedRandomAttachmentsContentCollector) getSourceWallPostsCount(source model.ContentSource) int {
	response, err := collector.client.WallGet(api.Params{
		"domain": source.VkCommunityID,
		"count":  1,
	})

	if err != nil {
		logrus.WithFields(packageLogFields).WithFields(logrus.Fields{
			"struct":   "CachedRandomAttachmentsContentCollector",
			"func":     "getSourceWallPostsCount",
			"err":      err,
			"fallback": "0 count",
		}).Error()
		return 0
	}

	return response.Count
}
