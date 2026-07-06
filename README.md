# GameFinder Bot

## Hosting

#### Compatible platforms
GameFinderBot has to be self hosted, but can run on virtually any server and Operating System. There is no released executable of the bot yet, so you will need to install [golang](https://go.dev/doc/install) to be able to build the executable yourself.

#### Specs requirements
GameFinderBot is lightweight and fast, if you only intend to use it on your own servers, any specs will do. Keep in mind however that it can at most support 30 servers simultaneously because of Helix API rate limits.

## Setup: Discord/Twitch Developer Portal 

### Discord
Before installing the bot, you need to create a Discord application and configure it properly.
 
1. Go to [Discord Developer Portal](https://discord.com/developers/applications) and create a **New Application**
2. In the **bot** section under **Privileged Gateway Intents**, enable:
   - **Server Members Intent**
   - **Message Content Intent**
3. Copy your bot token (**Reset Token** → copy). You will need it later.
4. Go to *OAuth2 → URL Generator*:
   - Check **bot** and **applications.commands**
   - Under Bot Permissions, check: **View Channels**, **Embed Links**
   - Copy the generated URL, open it in your browser, and invite the bot to your server.
5. Your `Application ID` is found under **General Information** in the developer portal.

### Twitch
1. Go to [Twitch Developers](https://dev.twitch.tv), log in with your twitch account (make one if you don't) and **register your application**
2. Give it a name, whatever you want.
3. For the **OAuth Redirection URL**, write `http://localhost`.
4. For the **Category**, select `Application Integration`
5. Lastly, note your **Client ID** and generate a new **Client Secret**, note it as well (you will only see it once).

## Installation (for Debian based-distros)

1. Install Go (Download the package directly from Go.dev instead of using `apt` as some older repositories won't download a recent enough version of Go) :
```bash
sudo apt remove golang-go -y # in case go is already installed
wget https://go.dev/dl/go1.26.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc # Adding Go to the environnement PATH
source ~/.bashrc
go version # should output something like "go version go1.26.4 linux/amd64"
```

2. Clone this repository to your local directory, and enter the folder :
```bash
cd ~
git clone https://github.com/RamenOsaka/game-finder-bot game-finder-bot
cd game-finder-bot
```

3. Install dependancies and compile the app :
```bash
go mod download
go build -o game-finder-bot # should create an executable called "game-finder-bot"
```

4. Set up the environnement variables `DISCORD_TOKEN`, `DISCORD_APP_ID`, `TWITCH_CLIENT_ID`  and `TWITCH_CLIENT_SECRET`.

You can create a `.env` file in the root folder of the bot and add one environnement variable by line with the synthax `VARIABLE=VALUE` if you do not want to add global environnenement variables.
* `DISCORD_TOKEN`is the identification token of your discord bot, it should be a 3 part alphanumerical string.
* `DISCORD_APP_ID` is a number associated to your application ID, found under **General Information** in the developer portal.
* `TWITCH_CLIENT_ID` is your twitch application Client ID, found in your application's informations.
* `TWITCH_CLIENT_SECRET` the identification token of your twitch bot that you generated earlier.

5. Lauching the application

If everything goes well, you can simply run the executable and the bot will be online and running. 
```bash
game-finder-bot
``` 
The application will create `config.json` if no file is found when the first config command is sent, this file stores every server configuration. If everything goes well, you will see `GameFinderBot is now running.  Press CTRL-C to exit.`.

## Running as a persistent service with systemd
 
To keep the bot running after closing your terminal and have it restart automatically on crash or reboot, set it up as a systemd service.
 
1. Create the service file
 
```bash
sudo nano /etc/systemd/system/game-finder-bot.service
```
 
Paste the following content (adjust `WorkingDirectory`, `User` and `ExecStart` if you cloned the repo to a different path):
 
```ini
[Unit]
Description=AutoBan Discord Bot (Go)
After=network.target
 
[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/game-finder-bot
ExecStart=/home/ubuntu/game-finder-bot/game-finder-bot
Restart=always
RestartSec=10
 
[Install]
WantedBy=multi-user.target
```
 
2. Enable and start the service
 
```bash
sudo systemctl daemon-reload
sudo systemctl enable game-finder-bot   # start automatically on reboot
sudo systemctl start game-finder-bot
sudo systemctl status game-finder-bot   # should show: Active: active (running)
```
 
# Useful service commands
 
| Command | Description |
|---|---|
| `sudo systemctl status game-finder-bot` | Check if the bot is running |
| `sudo systemctl restart game-finder-bot` | Restart the bot |
| `sudo systemctl stop game-finder-bot` | Stop the bot |
| `sudo journalctl -u game-finder-bot -f` | Watch live logs |
| `sudo journalctl -u game-finder-bot -n 50` | Show last 50 log lines |
 
# Updating the bot
 
When updates happen on GitHub, pull and rebuild on your server:
 
```bash
cd ~/game-finder-bot
git pull
go build -o game-finder-bot
sudo systemctl restart game-finder-bot
sudo systemctl status game-finder-bot
```
---
## Functionnalities
* Fetching and displaying twitch streams of a particular game in a selected channel.
* Filtering by minimum viewership.
* Adding streamers to a blacklist to stop them from showing up.
* Automatically deleting messages of streams which are no longer playing the game/are offline (can be disabled).

## Commands
* `/setlogchannel` sets the log channel, which is used to send messages if the bot encounters problems.
* `/setstreamchannel` sets the stream channel where twitch stream embeds will be displayed.
* `/setgame` sets the game to be displayed.
* `/startpolling` starts to poll the twitch API for games (and posting them in the stream channel).
* `/stoppolling` stops to poll the twitch API for games.
* `/addblacklist` adds a twitch user to the blacklist (works with login name and not display name).
* `/removeblacklist` removes a twitch user to the blacklist
* `/enableviewerfloor` enables or disables the minimum viewer cap.
* `/setminviewers` sets the viewer cap (set to 0 by default).
* `/enablehistory` enables or disables the stream message history. If enabled, stream embeds will never be deleted (disabled by default).
