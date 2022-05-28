package static

import "chattweiler/pkg/app/configs"

var VkCommunityBotToken = configs.NewMandatoryConfig("vk.community.bot.token")
var VkCommunityID = configs.NewMandatoryConfig("vk.community.id")
var VkCommunityChatID = configs.NewMandatoryConfig("vk.community.chat.id")
var VkAdminUserToken = configs.NewOptionalConfig("vk.admin.user.token", "")

var PgDatasourceString = configs.NewMandatoryConfig("pg.datasource.string")
var PgPhrasesCacheRefreshInterval = configs.NewOptionalConfig("pg.phrases.cache.refresh.interval", "15m")
var PgContentSourceCacheRefreshInterval = configs.NewOptionalConfig("pg.content.source.cache.refresh.interval", "15m")

var ChatWarderMembershipCheckInterval = configs.NewOptionalConfig("chat.warden.membership.check.interval", "10m")
var ChatWardenMembershipGracePeriod = configs.NewOptionalConfig("chat.warden.membership.grace.period", "1h")
var ChatUseFirstNameInsteadUsername = configs.NewOptionalConfig("chat.use.first.name.instead.username", "false")

var ContentAudioMaxCachedAttachments = configs.NewOptionalConfig("content.audio.max.cached.attachments", "100")
var ContentAudioCacheRefreshThreshold = configs.NewOptionalConfig("content.audio.cache.refresh.threshold", "0.2")
var ContentAudioQueueSize = configs.NewOptionalConfig("content.audio.queue.size", "100")
var ContentPictureMaxCachedAttachments = configs.NewOptionalConfig("content.picture.max.cached.attachments", "100")
var ContentPictureCacheRefreshThreshold = configs.NewOptionalConfig("content.picture.cache.refresh.threshold", "0.2")
var ContentPictureQueueSize = configs.NewOptionalConfig("content.picture.queue.size", "100")

var BotCommandOverrideInfo = configs.NewOptionalConfig("bot.command.override.info", "bark!")
var BotCommandOverrideAudioRequest = configs.NewOptionalConfig("bot.command.override.audio.request", "sing song!")
var BotCommandOverridePictureRequest = configs.NewOptionalConfig("bot.command.override.picture.request", "gimme pic!")
var BotFunctionalityWelcomeNewMembers = configs.NewOptionalConfig("bot.functionality.welcome.new.members", "true")
var BotFunctionalityGoodbyeMembers = configs.NewOptionalConfig("bot.functionality.goodbye.members", "true")
var BotFunctionalityMembershipChecking = configs.NewOptionalConfig("bot.functionality.membership.checking", "true")
var BotFunctionalityAudioRequests = configs.NewOptionalConfig("bot.functionality.audio.requests", "false")
var BotFunctionalityPictureRequests = configs.NewOptionalConfig("bot.functionality.picture.requests", "false")

var YandexObjectStorageAccessKeyID = configs.NewMandatoryConfig("yandex.object.storage.access.key.id")
var YandexObjectStorageSecretAccessKey = configs.NewMandatoryConfig("yandex.object.storage.secret.access.key")
var YandexObjectStorageRegion = configs.NewMandatoryConfig("yandex.object.storage.region")
var YandexObjectStoragePhrasesCacheRefreshInterval = configs.NewOptionalConfig("yandex.object.storage.phrases.cache.refresh.interval", "15m")
var YandexObjectStoragePhrasesBucket = configs.NewMandatoryConfig("yandex.object.storage.phrases.bucket")
var YandexObjectStoragePhrasesBucketKey = configs.NewMandatoryConfig("yandex.object.storage.phrases.bucket.key")
var YandexObjectStorageContentSourceCacheRefreshInterval = configs.NewOptionalConfig("yandex.object.storage.content.source.cache.refresh.interval", "15m")
var YandexObjectStorageContentSourceBucket = configs.NewMandatoryConfig("yandex.object.storage.content.source.bucket")
var YandexObjectStorageContentSourceBucketKey = configs.NewMandatoryConfig("yandex.object.storage.content.source.bucket.key")
var YandexObjectStorageMembershipWarningBucket = configs.NewMandatoryConfig("yandex.object.storage.membership.warning.bucket")
