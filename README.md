# 🌱 miku
> *Tiny, stateless microservice to notify that your Discord bot is going under maintenance, made in Go*

## Why?
Since many things can happen from your bot going offline, this is a tiny microservice
to say "Hi, <bot name> is undergoing maintenance, please wait!"

Since this is stateless, only when you invoke your bots prefix (which is configured using the `MIKU_PREFIX` environment variable),
it will check if it has been sent that message once and sends it to that specific channel and will not say it again

It will also add a "Do Not Disturb" presence with saying "Going under maintenance!"

## License
**miku** is released under the **MIT** License, read [here](/LICENSE) for more information in the [root repository](https://github.com/NinoDiscord/miku/blob/master/LICENSE)!
