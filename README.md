<div align="center">
    <img src="https://user-images.githubusercontent.com/44072343/155874103-b1757bd9-0b31-4e8c-8a74-bdf372f71ef5.png" width="250" height="200" alt="logo">
</div>

# Description

Chattweiler is a chat bot for Vkontankte. Briefly say, it takes care of a chat.

# Bot's Features

<details>
   <summary>Controls new chat joins: welcomes new members<</summary><br>
   
   You also can configure some audio with a welcome phrase. See "PostgreSQL Diagrams" for details
   
   <img src="https://user-images.githubusercontent.com/44072343/160234699-3699973e-615e-40eb-811a-a7a790cd8288.png" alt="hello new member example">
</details>

<details>
   <summary>Controls chat leavings: says goodbye</summary><br>
   <img src="https://user-images.githubusercontent.com/44072343/160234884-090b99e2-102e-43aa-ae0a-18bf8a66c191.png" alt="goodbye member example">
</details>

<details>
   <summary>Controls a membership at a chat's community: makes warnings and kicks if people don't care about it</summary><br>
   <img src="https://user-images.githubusercontent.com/44072343/160234979-5b19ee74-2be6-44a3-95eb-9193f2d38086.png" alt="warning for member example">
</details>

<details>
   <summary>Can send a random content to the chat</summary><br>
   
   The commands' names could be overridden. See "Configurations" for details
   
   <img src="https://user-images.githubusercontent.com/44072343/160235262-31cda1b1-e880-4c06-8676-edb1f00f598e.png" alt="picture command example"><br>
   <img src="https://user-images.githubusercontent.com/44072343/160235343-05e81966-a5b7-4658-b5b8-81565f4bf2a6.png" alt="audio command example">
</details>

# Development Stack

- Golang

- PostgreSQL

# PostgreSQL Diagrams

<img src="https://user-images.githubusercontent.com/44072343/160235542-063309c1-1d4e-46af-b8b3-7050a7f403ae.png" alt="diagrams">
<br>

<details>
    <summary>phrase_type</summary><br>
    
Phrases could have different types.

By default the application uses these types:

- `welcome`
- `goodbye`
- `membership_warning`
- `info`
- `audio_request`
- `picture_request`
    
</details>

<details>
    <summary>phrase</summary><br>
    
- `text` is an actual phrase
- `is_user_templated` means that the `text` can has inside a `%username%` mark which tells to the application to replace it to an actual username
- `weight` brings a bit of probability. Allows the application to choose a phrase by it's probability (counts only between phrases with the same phrase type). Details: <a href="https://en.wikipedia.org/wiki/Fitness_proportionate_selection">Fitness proportionate selection</a>
- `vk_audio_id` is an audio's id at Vkontakte, the application attaches it to a message if `is_audio_accompaniment` is true. Example of `vk_audio_id`, audio-2001545048_57545048

</details>

<details>
    <summary>membership_warning</summary><br>

Contains information about membership warnings.

- `first_warning_ts` is a timestamp which tells about when the first time a member was notified about a membership
- `grace_period` is a period which the application uses to define warning's status after `first_warning_ts`. For example, if `first_warning_ts + grace_period` less than `now()` then the warning has expired status
- `is_relevant` is a flag which tells about a current status of a warning. For example, if warning already sent and grace period still justified, it has true, otherwise it has false.

</details>

<details>
    <summary>source_type</summary><br>

Content sources could have different types.

By default the application uses these types:

- `audio`
- `picture`

</details>

<details>
    <summary>content_source</summary><br>

- `vk_community_id` is a url name of a community. Example, vk.com/awesome_community. Here awesome_community is the url name.

</details>

# Configurations

**Mandatory configurations**

|  Variable Name   | Description |
| -------------   | ------------- |
| `vk.community.bot.token (string)`     | You cant take it here *vk.com/<your_community>?act=tokens*  |
| `vk.community.id (int)`   | It should be a positive integer value. You can take it by a click on some post at a community and take it from the url, like *vk.com/<your_community>?w=wall-<community_id>_8851* |
| `vk.community.chat.id (int)`   | If you know peerId of a chat then a chat id will be like `peerId - 2000000000` result. Usually it has a sequence, so if it's your the first chat in a community then the chat id should be like 1 |
| `pg.datasource.string (string)` | Example, *"host=localhost user=postgres password=postgres sslmode=disable dbname=chattweiler"* or *"postgresql://username:password@host:port/dbname?param1=arg1"* |

<details>
    <summary>vk optional configurations</summary><br>

|  Variable Name | Default value | Description |
| ------------- | ------------- | ------------- |
| `vk.admin.user.token` | `"" (string)` | A community admin's "Implicit flow" token. Mandatory if you want to be able to use audio and picture request commands. <a href="https://dev.vk.com/api/access-token/getting-started">How to get it</a> |

</details>

<details>
    <summary>pg optional configurations</summary><br>

|  Variable Name | Default value | Description |
| ------------- | ------------- | ------------- |
| `pg.phrases.cache.refresh.interval` | `15m (string, golang type - time.Duration)` | A phrases cache refresh interval |
| `pg.content.source.cache.refresh.interval` | `15m (string, golang type - time.Duration)` | A content sources cache refresh interval |

</details>

<details>
    <summary>chat optional configurations</summary><br>

|  Variable Name  | Default value | Description |
| -------------  | ------------- | ------------- |
| `chat.warden.membership.check.interval` | `10m (string, golang type - time.Duration)` | An interval after which starts an async worker to check a chat for new membership warnings |
| `chat.warden.membership.grace.period` | `1h (string, golang type - time.Duration)` | A period that the application will assign to new warnings about a membership |
| `chat.use.first.name.instead.username` | `false (boolean)` | A toggle for using a first name of a member instead of his username. For example, Ammy (Joe etc.) instead of @username |

</details>

<details>
    <summary>content optional configurations</summary><br>

|  Variable Name  | Default value | Description |
| -------------  | ------------- | ------------- |
| `content.audio.max.cached.attachments` | `100 (int)` | A number of maximum available cached audio attachments |
| `content.audio.cache.refresh.threshold` | `0.2 (float)` | A float value between 0.0 and 1.0. Used for audios cache refreshing |
| `content.audio.queue.size` | `100 (int)` | A number of maximum requests queue for audio |
| `content.picture.max.cached.attachments` | `100 (int)` | A number of maximum available cached picture attachments |
| `content.picture.cache.refresh.threshold` | `0.2 (float)` | A float value between 0.0 and 1.0. Used for pictures cache refreshing |
| `content.picture.queue.size` | `100 (int)` | A number of maximum requests queue for picture |

</details>

<details>
    <summary>bot optional configurations</summary><br>

|  Variable Name   | Default value | Description |
| -------------  | ------------- | ------------- |
| `bot.command.override.info` | `bark! (string)` | A variable for info command name overriding |
| `bot.command.override.audio.request` | `sing song! (string)` | A variable for audio request command name overriding |
| `bot.command.override.picture.request` | `gimme pic! (string)` | A variable for picture request command name overriding |
| `bot.functionality.welcome.new.members` | `true (boolean)` | A toggle for new members welcome functions |
| `bot.functionality.goodbye.members` | `true (boolean)` | A toggle for goodbye members' leavings functions |
| `bot.functionality.membership.checking` | `true (boolean)` | A toggle for membership checking functions  |
| `bot.functionality.audio.requests` | `false (boolean)` | A toggle for audio requests handling functions |
| `bot.functionality.picture.requests` | `false (boolean)` | A toggle for picture requests handling functions |

</details>

# How To Run

Before the application will be running, create the file `bot.env` inside the project directory and fill it up by all the mandatory environment variables.

Run the commands in the console:

```
# optional if you have your own postgres running
# also see: sql/initdb.sql - initial scripts
./runPgLocally.sh 

# build a docker image and run a container using the image 
./build.sh && ./run.sh
```
