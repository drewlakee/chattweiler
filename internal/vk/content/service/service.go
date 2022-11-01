package service

import (
	"chattweiler/internal/configs"
	"chattweiler/internal/logging"
	"chattweiler/internal/utils"
	"chattweiler/internal/vk"
	"strconv"
)

var logPackage = "service"

func getMaxCachedAttachments(mediaType vk.MediaAttachmentType) int {
	switch mediaType {
	case vk.PhotoType:
		pictureMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentPictureMaxCachedAttachments), 10, 32)
		if err != nil {
			logging.Log.Panic(logPackage, "getMaxCachedAttachments", err, "%s: parsing of env variable is failed", configs.ContentPictureMaxCachedAttachments.Key)
		}

		return int(pictureMaxCachedAttachments)
	case vk.AudioType:
		audioMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentAudioMaxCachedAttachments), 10, 32)
		if err != nil {
			logging.Log.Panic(logPackage, "getMaxCachedAttachments", err, "%s: parsing of env variable is failed", configs.ContentAudioMaxCachedAttachments.Key)
		}

		return int(audioMaxCachedAttachments)
	case vk.VideoType:
		videoMaxCachedAttachments, err := strconv.ParseInt(utils.GetEnvOrDefault(configs.ContentVideoMaxCachedAttachments), 10, 32)
		if err != nil {
			logging.Log.Panic(logPackage, "getMaxCachedAttachments", err, "%s: parsing of env variable is failed", configs.ContentVideoMaxCachedAttachments.Key)
		}

		return int(videoMaxCachedAttachments)
	}

	return 0
}

func getCacheRefreshThresholdFor(mediaType vk.MediaAttachmentType) float32 {
	switch mediaType {
	case vk.PhotoType:
		pictureCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentPictureCacheRefreshThreshold), 32)
		if err != nil {
			logging.Log.Panic(logPackage, "getCacheRefreshThresholdFor", err, "%s: parsing of env variable is failed", configs.ContentPictureCacheRefreshThreshold.Key)
		}

		return float32(pictureCacheRefreshThreshold)
	case vk.AudioType:
		audioCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentAudioCacheRefreshThreshold), 32)
		if err != nil {
			logging.Log.Panic(logPackage, "getCacheRefreshThresholdFor", err, "%s: parsing of env variable is failed", configs.ContentAudioCacheRefreshThreshold.Key)
		}

		return float32(audioCacheRefreshThreshold)
	case vk.VideoType:
		videoCacheRefreshThreshold, err := strconv.ParseFloat(utils.GetEnvOrDefault(configs.ContentVideoCacheRefreshThreshold), 32)
		if err != nil {
			logging.Log.Panic(logPackage, "getCacheRefreshThresholdFor", err, "%s: parsing of env variable is failed", configs.ContentAudioCacheRefreshThreshold.Key)
		}

		return float32(videoCacheRefreshThreshold)
	}

	return 0
}
