package configs

/*
VkCommunityBotToken a specific token for your community (e.g. "956c94e96...6039be4e")
VkCommunityID a specific community id (e.g. "161...464")
VkCommunityChatID a particular chat in your community (e.g. either 1 or 2 or more)
VkAdminUserToken

Configurations for VK-API interactions
*/
var VkCommunityBotToken = NewMandatoryConfig("vk.community.bot.token")
var VkCommunityID = NewMandatoryConfig("vk.community.id")
var VkCommunityChatID = NewMandatoryConfig("vk.community.chat.id")
var VkAdminUserToken = NewOptionalConfig("vk.admin.user.token", "")

/*
ChatWarderMembershipCheckInterval a periodic interval after which the application goes to VK-API to compare actual members in a chat
ChatWardenMembershipGracePeriod a period after which the application checks if a warned user subscribed to a community
ChatUseFirstNameInsteadUsername either uses actual name of a user or his url-uid for communication (e.g. "John" or "john_2001")

Configurations for the chat functionality
*/
var ChatWarderMembershipCheckInterval = NewOptionalConfig("chat.warden.membership.check.interval", "10m")
var ChatWardenMembershipGracePeriod = NewOptionalConfig("chat.warden.membership.grace.period", "1h")
var ChatUseFirstNameInsteadUsername = NewOptionalConfig("chat.use.first.name.instead.username", "false")

/*
ContentCommandCacheRefreshInterval a periodic interval after which the application invalidates its cache with commands
ContentRequestsQueueSize a buffered channel size between event handler and command executors
ContentGarbageCollectorsCleaningInterval a periodic interval after which the application removes already unused content collectors which are cached

Configurations for content commands` logic
*/
var ContentCommandCacheRefreshInterval = NewOptionalConfig("content.command.cache.refresh.interval", "15m")
var ContentRequestsQueueSize = NewOptionalConfig("content.requests.queue.size", "100")
var ContentGarbageCollectorsCleaningInterval = NewOptionalConfig("content.garbage.collectors.cleaning.interval", "10m")

// PhrasesCacheRefreshInterval a periodic interval after which the application invalidates its cache with phrases
var PhrasesCacheRefreshInterval = NewOptionalConfig("phrases.cache.refresh.interval", "15m")

// ContentAudioMaxCachedAttachments a max number of content that could be stored in an application's cache
// ContentAudioCacheRefreshThreshold a threshold for a cache with content after which the cache fills out by new content
var ContentAudioMaxCachedAttachments = NewOptionalConfig("content.audio.max.cached.attachments", "100")
var ContentAudioCacheRefreshThreshold = NewOptionalConfig("content.audio.cache.refresh.threshold", "0.2")

// ContentPictureMaxCachedAttachments a max number of content that could be stored in an application's cache
// ContentPictureCacheRefreshThreshold a threshold for a cache with content after which the cache fills out by new content
var ContentPictureMaxCachedAttachments = NewOptionalConfig("content.picture.max.cached.attachments", "100")
var ContentPictureCacheRefreshThreshold = NewOptionalConfig("content.picture.cache.refresh.threshold", "0.2")

// ContentVideoMaxCachedAttachments a max number of content that could be stored in an application's cache
// ContentVideoCacheRefreshThreshold a threshold for a cache with content after which the cache fills out by new content
var ContentVideoMaxCachedAttachments = NewOptionalConfig("content.video.max.cached.attachments", "100")
var ContentVideoCacheRefreshThreshold = NewOptionalConfig("content.video.cache.refresh.threshold", "0.2")

/*
BotFunctionalityWelcomeNewMembers enables welcome functionality
BotFunctionalityGoodbyeMembers enables goodbye functionality
BotFunctionalityMembershipChecking enables membership checking functionality
BotFunctionalityContentCommands enables requesting of media content functionality
BotLogToFile enables writing of a log file near an execution file

General application configurations
*/
var BotFunctionalityWelcomeNewMembers = NewOptionalConfig("bot.functionality.welcome.new.members", "true")
var BotFunctionalityGoodbyeMembers = NewOptionalConfig("bot.functionality.goodbye.members", "true")
var BotFunctionalityMembershipChecking = NewOptionalConfig("bot.functionality.membership.checking", "false")
var BotFunctionalityContentCommands = NewOptionalConfig("bot.functionality.content.commands", "false")
var BotLogToFile = NewOptionalConfig("bot.log.file", "false")

/*
YandexObjectStorageAccessKeyID
YandexObjectStorageSecretAccessKey
YandexObjectStorageRegion
YandexObjectStoragePhrasesBucket
YandexObjectStoragePhrasesBucketKey
YandexObjectStorageContentSourceBucket
YandexObjectStorageContentSourceBucketKey
YandexObjectStorageMembershipWarningBucket

https://cloud.yandex.com/en-ru/services/storage
Yandex S3 object storage configurations
*/
var YandexObjectStorageAccessKeyID = NewMandatoryConfig("yandex.object.storage.access.key.id")
var YandexObjectStorageSecretAccessKey = NewMandatoryConfig("yandex.object.storage.secret.access.key")
var YandexObjectStorageRegion = NewMandatoryConfig("yandex.object.storage.region")
var YandexObjectStoragePhrasesBucket = NewMandatoryConfig("yandex.object.storage.phrases.bucket")
var YandexObjectStoragePhrasesBucketKey = NewMandatoryConfig("yandex.object.storage.phrases.bucket.key")
var YandexObjectStorageContentSourceBucket = NewMandatoryConfig("yandex.object.storage.content.command.bucket")
var YandexObjectStorageContentSourceBucketKey = NewMandatoryConfig("yandex.object.storage.content.command.bucket.key")
var YandexObjectStorageMembershipWarningBucket = NewMandatoryConfig("yandex.object.storage.membership.warning.bucket")
