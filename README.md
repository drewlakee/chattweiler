# Chattweiler

Chattweiler is a server-side application that handles events which come from VK community chat

Inspirations for the development are:
- Requesting of content from custom sources
- Customized responses
- Fake users filtering
- Forcing to join a community, in case, if you're wanting to be a part of a chat
- Fun 🤖

## Features

- Customized chat responses for chat joins, leavings, warnings, failed commands etc.
- Customized commands for content (e.g. pictures, audio, videos)
- Automatic membership checking and warning

## Application context schema

<div align="center">
    <img src="https://user-images.githubusercontent.com/44072343/200119060-c7feda44-b3ca-40ab-afb5-23cef2fddc8a.jpg" alt="logo">
</div>

As you might already have noticed the application uses for storage [Yandex Object Storage](https://cloud.yandex.com/en-ru/services/storage) solution, 
so for using the application you have to have access to such resource. Storage configuration for the application is mentioned further in Quickstart.

In brief, the application operates over csv files that are stored in cloud. That type of file is picked up because it's very straightforward to store and edit.
The application caches these files and invalidates over time. That way makes positive effect on performance during events handling.

# Quickstart
## Application deployment preparations
### Setting up a community chat

1. Create a public community in [VK](https://vk.com)
2. Create a new chat inside the community: Manage > Chats > Create Chat
3. Once you got to a page of the chat, remember the chat's number (e.g. 9)

![Screenshot 2022-10-31 165551](https://user-images.githubusercontent.com/44072343/199024841-a4da7cb9-829d-43ed-9abc-df60378b124f.png)

### Setting up a Yandex Object Storage

The application uses several [buckets](https://console.cloud.yandex.com/folders): 

- commands
  - commands_production.csv
- membership-warnings (optional, used automatically by the application if so configured)
- phrases
  - phrases_production.csv

For further configurations you have to have such buckets in your environment.

#### Phrases

File must contain rows with a specific structure:

```go
type PhraseType string

const (
	// for new users in chat
	WelcomeType           PhraseType = "welcome"
	// for users who left
	GoodbyeType           PhraseType = "goodbye"
	// for users who in chat but not in a community
	MembershipWarningType PhraseType = "membership_warning"
	// for some general info like commands description
	InfoType              PhraseType = "info"
	// for responses with content requests
	ContentRequestType    PhraseType = "content_request" 
	// for cases where the application failed to find something
	RetryType             PhraseType = "retry_request"
)

type Phrase struct {
	PhraseID   int        `csv:"phrase_id"`
	// used for probability
	// https://en.wikipedia.org/wiki/Fitness_proportionate_selection
	// in brief, if its value more than others` value it has more chances to be picked up
	Weight     int        `csv:"weight"`
	PhraseType PhraseType `csv:"phrase_type"`
	// use null if command is not supposed to use it
	VkAudioId  string     `csv:"vk_audio_id"`
	// use null if command is not supposed to use it
	VkGifId    string     `csv:"vk_gif_id"`
	// actual text of a phrase
	Text       string     `csv:"text"`
}
```
```
csv file:

phrase_id1,weight,phrase_type,vk_audio_id,vk_gif_id,text
phrase_id2,weight,phrase_type,vk_audio_id,vk_gif_id,text
...
1,100,welcome,null,doc120747496_641221964,"Hello there, %username%!"
5,100,membership_warning,null,doc120747496_641228085,"%username%, this chat is only for community members 👻\nPlease subscribe quickly!"
18,100,retry_request,null,doc120747496_646353718,"%username%, oops, we've failed, try again 👉🏻👈🏻"
```

Phrases are used for responses on different types of events.

#### Commands

File must contain rows with a specific structure: 

```go
type CommandType string

const (
	InfoCommand    CommandType = "info"
	ContentCommand CommandType = "content"
)

// CsvCommand storage specific object of Command
type CsvCommand struct {
	ID                int         `csv:"id"`
	Commands          string      `csv:"commands"`
	Type              CommandType `csv:"command_type"`
	MediaContentTypes string      `csv:"media_types"`
	CommunityIDs      string      `csv:"community_ids"`
}
```

```
csv file:

id1,"alias1,alias2",command_type,"media_type1,media_type2","community_id1,community_id2"
id2,"alias1,alias2",command_type,"media_type1,media_type2","community_id1,community_id2"
...
6,"jazzy music,🥸",content,audio,jazzjazz
26,"👾,commands",info,,
```

- A command could have several aliases which users can call on in chat

- A command can use several communities to fetch content from it

- A command can has several media-content types to fetch from communities (randomly chosen per call)

- `command_type` used for different types of command. There's a couple of them right now, command with `info` type sends in chat a phrase with the same type 

#### Membership warnings

```go
type MembershipWarning struct {
	WarningID      int       `csv:"warning_id"`
	UserID         int       `csv:"user_id"`
	Username       string    `csv:"username"`
	// when user got first warning in chat about community membership
	FirstWarningTs time.Time `csv:"first_warning_ts"`
	// a period in which he has to subscribe, or he'll be kicked eventually
	GracePeriod    string    `csv:"grace_period"`
	// actual status of a warning 
	// if a user got a warning and subscribed, then status will be updated
	IsRelevant     bool      `csv:"is_relevant"`
}
```

If you want to use such feature, then that structure will be used to upload actual status about warnings to a storage bucket by days.

- 2022-23-10
- 2022-24-10
- 2022-25-10
- .....

Such files occur only if warnings happen in a day, so there could be some gaps between files.

## Local application deployment

### Application configurations

**Mandatory configurations**

- `vk.community.bot.token`

A specific token for your community (e.g. "956c94e96...6039be4e")

How to get: Enter your community > Manage > Settings > API usage > Access tokens

- `vk.community.id`

A specific community id (e.g. "161...464" as a number)

You can get it somewhere in a community or by picking up from some wallpost's url `https://vk.com/community?w=wall-<id>_3394`

- `vk.community.chat.id`

An actual number of a chat, we've mentioned it earlier in the Quickstart

- Yandex Object Storage
  - `yandex.object.storage.access.key.id` (e.g. some token like `YCN1Ze...SJv`)
  - `yandex.object.storage.secret.access.key` (e.g. some token like `YCA...cQ`)
  - `yandex.object.storage.region` (e.g. `ru-central1`)
  - `yandex.object.storage.phrases.bucket` (e.g. `phrases-bucket`)
  - `yandex.object.storage.phrases.bucket.key` (e.g. `phrases_production.csv`)
  - `yandex.object.storage.content.command.bucket` (e.g. `command-bucket`)
  - `yandex.object.storage.content.command.bucket.key` (e.g `command_production.csv`)
  - `yandex.object.storage.membership.warning.bucket` (e.g. `membership-warning-bucket`)

Read [the documentation](https://cloud.yandex.com/en-ru/docs/storage/) how to get these values

**Optional configurations**

- `vk.admin.user.token` (by default not specified) if you're supposed to use content requesting, you have to have that one. Read [the documentation](https://dev.vk.com/api/access-token/implicit-flow-user) how to get such token

- `chat.warden.membership.check.interval` (default: `10m`) a periodic interval after which the application goes to VK-API to compare actual members in a chat
- `chat.warden.membership.grace.period` (default: `1h`) a period after which the application checks if a warned user subscribed to a community
- `chat.use.first.name.instead.username` (default: `false`) either uses actual name of a user or his url-uid for communication (e.g. "John" or "john_2001")
- `content.command.cache.refresh.interval` (default: `15m`) a periodic interval after which the application invalidates its cache with commands
- `content.requests.queue.size` (default: `100`) a buffered channel size between event handler and command executors
- `content.garbage.collectors.cleaning.interval` (default: `10m`) a periodic interval after which the application removes already unused content collectors which are cached
- `phrases.cache.refresh.interval` (default: `15m`) a periodic interval after which the application invalidates its cache with phrases
- `content.audio.max.cached.attachments` (default: `100`) a max number of content that could be stored in an application's cache
- `content.audio.cache.refresh.threshold` (default: `0.2`) a threshold for a cache with content after which the cache fills out by new content
- `content.picture.max.cached.attachments` (default: `100`) a max number of content that could be stored in an application's cache
- `content.picture.cache.refresh.threshold` (default: `0.2`) a threshold for a cache with content after which the cache fills out by new content
- `content.video.max.cached.attachments` (default: `100`) a max number of content that could be stored in an application's cache
- `content.video.cache.refresh.threshold` (default: `0.2`) a threshold for a cache with content after which the cache fills out by new content
- `bot.functionality.welcome.new.members` (default: `true`) enables welcome functionality
- `bot.functionality.goodbye.members` (default: `true`) enables goodbye functionality
- `bot.functionality.membership.checking` (default: `false`) enables membership checking functionality
- `bot.functionality.content.commands` (default: `false`) enables requesting of media content functionality
- `bot.log.file` (default: `false`) enables writing of a log file near an execution file

### Deployment

1. Clone the project `git clone git@github.com:drewlakee/chattweiler.git`
2. Build a docker image `./chattweiler/build.sh`
3. Create a configuration file `touch bot.env` and fill the mandatory variables
4. Run a container with the image you've just built `./chattweiler/run.sh`
5. Make fun out of it 👾

<details>
  <summary><b>Usage examples</b></summary>
  
![Screenshot 2022-10-31 at 16-11-04 Messenger](https://user-images.githubusercontent.com/44072343/199244389-1d16c36d-5136-4223-b8c8-959e29da4aeb.png)
  
![Screenshot 2022-10-31 at 16-13-06 Messenger](https://user-images.githubusercontent.com/44072343/199244378-b49e6aa0-7d94-41a7-b723-da94ed4d7ec5.png)

</details>

** If you are supposed to use file logging, you can make a volume by adding to the command in `./chattweiler/run.sh` a piece of settings `docker run -v /path/to/your/log/directory:/application/logs ...`
