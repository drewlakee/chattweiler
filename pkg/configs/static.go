package configs

var VkCommunityBotToken = NewMandatoryConfig("vk.community.bot.token")
var VkCommunityID = NewMandatoryConfig("vk.community.id")
var VkCommunityChatID = NewMandatoryConfig("vk.community.chat.id")
var VkAdminUserToken = NewOptionalConfig("vk.admin.user.token", "")

var PgDatasourceString = NewMandatoryConfig("pg.datasource.string")

var ChatWarderMembershipCheckInterval = NewOptionalConfig("chat.warden.membership.check.interval", "10m")
var ChatWardenMembershipGracePeriod = NewOptionalConfig("chat.warden.membership.grace.period", "1h")
var ChatUseFirstNameInsteadUsername = NewOptionalConfig("chat.use.first.name.instead.username", "false")

var PhrasesCacheRefreshInterval = NewOptionalConfig("phrases.cache.refresh.interval", "15m")

var ContentSourceCacheRefreshInterval = NewOptionalConfig("content.source.cache.refresh.interval", "15m")
var ContentAudioMaxCachedAttachments = NewOptionalConfig("content.audio.max.cached.attachments", "100")
var ContentAudioCacheRefreshThreshold = NewOptionalConfig("content.audio.cache.refresh.threshold", "0.2")
var ContentAudioQueueSize = NewOptionalConfig("content.audio.queue.size", "100")
var ContentPictureMaxCachedAttachments = NewOptionalConfig("content.picture.max.cached.attachments", "100")
var ContentPictureCacheRefreshThreshold = NewOptionalConfig("content.picture.cache.refresh.threshold", "0.2")
var ContentPictureQueueSize = NewOptionalConfig("content.picture.queue.size", "100")

var BotCommandOverrideInfo = NewOptionalConfig("bot.command.override.info", "bark!")
var BotCommandOverrideAudioRequest = NewOptionalConfig("bot.command.override.audio.request", "sing song!")
var BotCommandOverridePictureRequest = NewOptionalConfig("bot.command.override.picture.request", "gimme pic!")
var BotFunctionalityWelcomeNewMembers = NewOptionalConfig("bot.functionality.welcome.new.members", "true")
var BotFunctionalityGoodbyeMembers = NewOptionalConfig("bot.functionality.goodbye.members", "true")
var BotFunctionalityMembershipChecking = NewOptionalConfig("bot.functionality.membership.checking", "true")
var BotFunctionalityAudioRequests = NewOptionalConfig("bot.functionality.audio.requests", "false")
var BotFunctionalityPictureRequests = NewOptionalConfig("bot.functionality.picture.requests", "false")

var YandexObjectStorageAccessKeyID = NewMandatoryConfig("yandex.object.storage.access.key.id")
var YandexObjectStorageSecretAccessKey = NewMandatoryConfig("yandex.object.storage.secret.access.key")
var YandexObjectStorageRegion = NewMandatoryConfig("yandex.object.storage.region")
var YandexObjectStoragePhrasesBucket = NewMandatoryConfig("yandex.object.storage.phrases.bucket")
var YandexObjectStoragePhrasesBucketKey = NewMandatoryConfig("yandex.object.storage.phrases.bucket.key")
var YandexObjectStorageContentSourceBucket = NewMandatoryConfig("yandex.object.storage.content.source.bucket")
var YandexObjectStorageContentSourceBucketKey = NewMandatoryConfig("yandex.object.storage.content.source.bucket.key")
var YandexObjectStorageMembershipWarningBucket = NewMandatoryConfig("yandex.object.storage.membership.warning.bucket")
