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
var PhrasesSuppressLogsMissedPhrases = NewOptionalConfig("phrases.suppress.logs.missed.phrases", "false")

var ContentCommandCacheRefreshInterval = NewOptionalConfig("content.command.cache.refresh.interval", "15m")
var ContentRequestsQueueSize = NewOptionalConfig("content.requests.queue.size", "100")
var ContentGarbageCollectorsCleaningInterval = NewOptionalConfig("content.garbage.collectors.cleaning.interval", "10m")

var ContentAudioMaxCachedAttachments = NewOptionalConfig("content.audio.max.cached.attachments", "100")
var ContentAudioCacheRefreshThreshold = NewOptionalConfig("content.audio.cache.refresh.threshold", "0.2")
var ContentPictureMaxCachedAttachments = NewOptionalConfig("content.picture.max.cached.attachments", "100")
var ContentPictureCacheRefreshThreshold = NewOptionalConfig("content.picture.cache.refresh.threshold", "0.2")
var ContentPictureQueueSize = NewOptionalConfig("content.picture.queue.size", "100")

var BotCommandOverrideInfo = NewOptionalConfig("bot.command.override.info", "info")
var BotFunctionalityWelcomeNewMembers = NewOptionalConfig("bot.functionality.welcome.new.members", "true")
var BotFunctionalityGoodbyeMembers = NewOptionalConfig("bot.functionality.goodbye.members", "true")
var BotFunctionalityMembershipChecking = NewOptionalConfig("bot.functionality.membership.checking", "false")
var BotFunctionalityContentCommands = NewOptionalConfig("bot.functionality.content.commands", "false")

var YandexObjectStorageAccessKeyID = NewMandatoryConfig("yandex.object.storage.access.key.id")
var YandexObjectStorageSecretAccessKey = NewMandatoryConfig("yandex.object.storage.secret.access.key")
var YandexObjectStorageRegion = NewMandatoryConfig("yandex.object.storage.region")
var YandexObjectStoragePhrasesBucket = NewMandatoryConfig("yandex.object.storage.phrases.bucket")
var YandexObjectStoragePhrasesBucketKey = NewMandatoryConfig("yandex.object.storage.phrases.bucket.key")
var YandexObjectStorageContentSourceBucket = NewMandatoryConfig("yandex.object.storage.content.command.bucket")
var YandexObjectStorageContentSourceBucketKey = NewMandatoryConfig("yandex.object.storage.content.command.bucket.key")
var YandexObjectStorageMembershipWarningBucket = NewMandatoryConfig("yandex.object.storage.membership.warning.bucket")
