<div align="center">
    <img src="https://user-images.githubusercontent.com/44072343/155874103-b1757bd9-0b31-4e8c-8a74-bdf372f71ef5.png" width="250" height="200" alt="logo">
</div>

# Description

Chattweiler is a chat bot for Vkontankte. Briefly say, it takes care of the chat.

# What Actually It Does

- Controls new chat joins: the bot welcomes all the new members
- Controls chat leavings: the bot says goodbye
- Controls membership at chat's community: the bot makes warnings, kicks if people don't care about their warnings

# Stack

- Golang
- PostgreSQL

# PostgreSQL diagrams

<img src="https://user-images.githubusercontent.com/44072343/155875606-e10b2ba4-94e2-4fd0-9609-7aa416785e86.png" width="500" height="400" alt="diagrams">
<br>

- `phrase_type` table: phrases can has different types. For example, `goodbye` type or `welcome` type, etc
- `phrase` table:
  - `text` is an actual phrase
  - `is_user_templated` means that a `text` can has inside a `%username%` mark, which tells to the application to replace it to an actual username
  - `weight` brings a bit of probability. Allows the application to choose a phrase by it's probability (takes account only between the same phrase type). Details: <a href="https://en.wikipedia.org/wiki/Fitness_proportionate_selection">Fitness proportionate selection</a>
  - `vk_audio_id` is an audio file id at Vkontakte, application adds it to a message-phrase if `is_audio_accompaniment` is true. Example, audio-2001545048_57545048
- `membership_warning` table: stores all warnings about a membership 
   - `grace_period` is a period, which the application uses to define expired warnings after `first_warning_ts`: if `first_warning_ts + grace_period` < `now()` then a warning is expired  

# Environment Variables

- vk
  - ```vk.community.bot.token``` (**mandatory**, type `string`): you cant take it here *vk.com/<your_community>?act=tokens*
  - ```vk.community.id``` (**mandatory**, type `int`): it should be an positive integer. You can take it by a click on some post at the community and take it from url, like *vk.com/<your_community>?w=wall-<community_id>_8851*
  - ```vk.community.chat.id``` (**mandatory**, type `int`): if you know peerId of a chat then the chat id will be like `peerId - 2000000000` result. Usually, it has a sequence, so if it's your the first chat in the community then chat id should be like 1
- pg
  - ```pg.datasource.string``` (**mandatory**, type `string`)
     - Example, `"host=localhost user=postgres password=postgres sslmode=disable dbname=chattweiler"` or `"postgresql://username:password@host:port/dbname?param1=arg1"`
  - ```pg.phrases.cache.refresh.interval``` (optional, type `time.Duration`, default `15m`): cache refresh interval
- chat
  - ```chat.warden.membership.check.interval``` (optional, type `time.Duration`, default `10m`): interval, which after the application launch starts an async worker to check a chat for new membership warnings
  - ```chat.warden.membership.grace.period``` (optional, type `time.Duration`, default `1h`): period that the application will assign to new warnings about membership
- bot
  - ```bot.functionality.welcome.new.members``` (optional, type `boolean`, default `true`): toggle for new members welcome functions
  - ```bot.functionality.goodbye.members``` (optional, type `boolean`, default `true`): toggle for goodbye members' leavings functions
  - ```bot.functionality.membership.checking``` (optional, type `boolean`, default `true`): toggle for membership checking functions 
  
# How To Run

Before the application will be running, create `bot.env` file inside the project directory and fill it up by all the mandatory environment variables

Run the commands in the console:

```
# optional if you have your own postgres running
./runPgLocally.sh 

# build a docker image and run a container using the image 
./build.sh && ./run.sh
```