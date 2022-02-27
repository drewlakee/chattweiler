<div align="center">
    <img src="https://user-images.githubusercontent.com/44072343/155874103-b1757bd9-0b31-4e8c-8a74-bdf372f71ef5.png" width="250" height="200" alt="logo">
</div>

# Description

Chattweiler is a chat bot for Vkontankte. Briefly say, it takes care of the chat.

# What Actually It Does

- Controls new chat joins: it welcomes all the new members
- Controls chat leavings: it says goodbye
- Controls membership at chat's community: makes warnings, kicks if people don't care

# Stack

- Golang application: all the logic
- PostgreSQL: stores all personal phrases and warnings

# PostgreSQL diagrams

<img src="https://user-images.githubusercontent.com/44072343/155875606-e10b2ba4-94e2-4fd0-9609-7aa416785e86.png" width="500" height="400" alt="diagrams">
<br>

- `phrase_type` table: phrases can has different type. For example, goodbye type or welcome etc.
- `phrase` table
  - `text` is an actual phrase
  - `is_user_templated` means that `text` can has inside `%username%` mark which tells to app replace it to an actual username
  - `weight` brings a bit of probability. Allows the application to choose a phrase by it's probability (takes account only between the same phrase type). Details: <a href="https://en.wikipedia.org/wiki/Fitness_proportionate_selection">Fitness proportionate selection</a>
  - `vk_audio_id` application adds it to message-phrase if  `is_audio_accompaniment` is true. Example, audio-2001545048_57545048
- `membership_warning` table: stores all the warnings about membership 
   - `grace_period` is a period which application uses to define expired warnings after `first_warning_ts `: if `first_warning_ts + grace_period` < `now()` then expired warning  

# Environment Variables

- vk
  - `vk.community.bot.token` (mandatory): You cant take it here *vk.com/<your_community>?act=tokens*
  - `vk.community.id` (mandatory): It should be an positive integer. You can take it by click on some post inside the community and take it from url, like *vk.com/<your_community>?w=wall-<community_id>_8851*
  - `vk.community.chat.id` (mandatory): If you know peerId of the chat then chat id will be `peerId - 2000000000`. Usually, it has a sequence, so if it's your first chat in the community then chat id should be 1
- pg
  - `pg.datasource.string` (mandatory)
     - Example, "host=localhost user=postgres password=postgres sslmode=disable dbname=chattweiler" or "postgresql://username:password@host:port/dbname?param1=arg1"
  - `pg.phrases.cache.refresh.interval` (mandatory): cache refresh interval, under the hood it uses golang type `time.Duration`, so input should be the same like 1h or 30s, etc.
- chat
  - `chat.warden.membership.check.interval` (mandatory): period/interval which after application will wake up it's async worker to check chat for new membership warnings
  - `chat.warden.membership.grace.period` (mandatory): period that application will assign to new warnings about membership
  
# How To Run

Before the application will run create `bot.env` file inside project directory and fill it up by all the mandatory environment variables

Run the commands in the console:

```
# optional if you have your own postgres running
./runPgLocally.sh 

# build a docker image and run a container using the image 
./build.sh && ./run.sh
```