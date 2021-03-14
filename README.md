
# SSTB - Super Simple Trading Bot

[![Latest Version on Packagist][ico-version]][link-repo]
[![Software License][ico-license]][link-license]
[![Total Downloads][ico-downloads]][link-repo]
[![Hits][ico-hits]][link-hits]

**!! IMPORTANT NOTICE !!**
>I don't take any responsibility for any loss or problems you may experience. You'll use 
>this software and any versions of it at your own risk.

## Description
This is a simple trading bot created just for the fun of it and to automate a simple trading 
strategy I used to do manually...
This bot is not well written, but works well for my use case.
 

## Table of Contents
- [Features](#features)
- [Requirements](#requirements)
- [Supported Exchanges](#supported-exchange-platforms)
- [Introduction](#introduction)
- [How does it work?](#how-does-it-work)
- [Questions & Answers](#questions--answers)
- [Quick guide](#getting-started---quick-guide)
- [Step-by-step Guide](#getting-started---step-by-step)
- [Dependencies](#dependencies)
- [Build](#build)
- [Service setup](#service-setup)
- [Arguments](#arguments)
- [Configuration](#configuration)
    - [App](#configuration)
        - [Provider](#provider)
        - [Notifier](#notifier)
    - [Jobs](#jobs)
    - [Logging](#logging)
- [Known issues](#known-issues)
- [Support](#support)
- [Donate](#donate)
- [Security](#security)
- [Credits](#credits)
- [License](#license)


## Features
- Automatic trading
- Sends notifications via email, [mattermost](https://github.com/mattermost/mattermost-server) or [slack](https://www.slack.com)


## Requirements:
None really..

##### Tested under:
- Debian 10
- Debian 9
- Ubuntu 20.04
- Ubuntu 18.04

##### Not tested (might still work):
- OSX
- Windows


## Supported exchange platforms
- **Binance** 
Use this [link](https://www.binance.com/en/register?ref=KLK9HBCF) or the code `KLK9HBCF` if 
you sign up to get 10% of all payed spot trade fees payed back by Binance.

- **Poloniex** 
Use this [link](https://poloniex.com/signup?c=4EJJK4JR) or the code `4EJJK4JR` if you sign 
up. This will support the future development of this bot.


## Introduction
Prepare yourself for a short reading lesson (10-15 min) and make sure you understand how the bot
operates before you actually use it. If you have any questions please feel free to create a 
new issue. If you have more than one question, just create several new issues :)

I've tried to explain the mechanics and used strategy as detailed as possible, but I'm sure I've
missed something. It's also very likely that you encounter a bug. The bot is written for my
own special use case and might behave differently if used in another context.

Please be careful and only use "play money". I'll wish you the best of luck and happy trading :)


## How does it work?
This bot will monitor all user orders for a given number of markets and will automatically 
create counter orders whenever one of these monitored orders gets fulfilled. 

Example: 
If you've sold 100 coins for a price of 0.03 BTC, the bot will create a buy order of 100 
coins for a price 0.02 BTC (given your job step size is set to 0.01). 
The same will happen to the opposite situation. If you've bought some coins, the bot will 
automatically sell them for more.
By doing so the bot will generate revenue as long as he is able to trade (buy cheap, 
sell for more).
This works especially well with coins traded below 200 Satoshi and with a step size of 
`0.00000001`. If the market moves up or down - it doesn't matter. You'll always have something to
trade and therefor generate revenue.


## Questions & Answers
##### Q: Is a profit guaranteed?
No! If the coin you invested in drops to zero, gets delisted, doesn't get traded enough, etc your 
investment is gone or isn't really increasing. Only use this bot if you know what you are doing!

##### Q: How do I know the source code doesn't include a "virus"?
I recommend you do a code review and compile the binary yourself. This way you can be sure no one 
has tampered with the code or the executable. 

##### Q: How many accounts can be used / markets be traded on?
There are no limitations. The bot will spawn a new thread for every job you provide. So it 
totally depends on your system. A few hundred should not be a problem...

##### Q: Which market is the best?
Well this is tough to say, look for a market which:
- is traded below 200 satoshi (0.000002 BTC) (if your trade step is set to 1 satoshi)
- has a high volatility (goes up and down a lot)
- gets traded enough to fulfill your trades
- looks good to you ;)

##### Q: Which step size should I choose?
Your step size can be as low as the exchange allows. However keep in mind that a trade has to
make at least 0.2% profit on Binance and 0.3% on Poloniex. If the profit falls below it, you will
loose coins on every trade.
You may also verify if your step size is profitable by looking at the current chart - would 
any trades be executed in the last 12/24/36/48 hours? 

##### Q: How much should I at least invest?
This depend on the coin you are interested in. This can be roughly calculated by answering the 
following questions:

- Currently traded value (Lets assume it is traded for 100 satoshi)
- Lowest expected value (maybe down to 30 satoshi?)
- Highest expected value (maybe up to 130 satoshi?)
- How frequently do I want to trade? A higher trade step will result in a higher win rate, but in a slower trading speed
- How high should the min trade volume be? (Many pairs have a trade minimum of 0.0001 BTC)

Total invest = ((((Currently traded value) - (Lowest expected value)) + ((Highest expected value) - (Currently traded value))) / (trade step)) * (trade volume)

Example:
- Currently traded value: `0.00000066`
- Lowest expected value: `0.00000030`
- Highest expected value: : `0.00000102`
- Trade step size: : `0.00000001`
- Trade volume: : `0.000101`

```
Total invest = (((0.00000066 - 0.00000030) + (0.00000102 - 0.00000066)) / 0.00000001) * 0.000101
Total invest = 0.007272 BTC
```

But this is only partially true. The sell side will cost slightly less, since you are able to 
buy them for less. However the sum above will represent the final investment if the bot gets
sold out - means there are no sell orders left to fulfill.

##### Q: Can the ROI be calculated?
The return of invest or short ROI can not be calculated. It totally depends on the volatility
of the market and your chosen parameters. 

You can try to get a sense by manually checking how many trades could have been made in lets say the
past 24 hours.
Now for each order you calculate the gain: 
```
buy total = (buy price) * (amount)
buy fee = ((buy total) / 100) * 0.15

sell total = (sell price) * (amount - (buy fee))
sell fee = ((sell total) / 100) * 0.15

gain = ((sell total) - (buy total)) - (sell fee)

percent = (gain / (buy total)) * 100
```
..now sum up each gain. The result is a rough estimate of the profit that could have been made. 

##### Q: Is this a "smart" bot?
No. This bot is just executing orders every time an other order got fulfilled. I won't consider
this "smart" behavior. It's just a strategy to gain profit through constant trading.. 

##### Q: My exchange is not supported.
If you would like to use the bot with a different exchange, feel free to create a new issue / 
feature request.

Please add the following information:
- Name & Website
- Link to the official API documentation

##### Q: Where is the GUI / How do I interact with the bot?
There is no gui required. Everything is already provided by the exchange. 
Are you missing anything specific? Feel free to create a new issue.

##### Q: How long do I have to wait? Nothing is happening!
Well, that's normal - the market isn't always super excited. Have some patience and if you are not
happy with the performance, perhaps another market would perform better?

##### Q: Why is my overall BTC balance decreasing?
The overall balance isn't that interesting, you should actually see rather high fluctuations since
the value of the coin you've invested in may decrease or increase. If you want to know how much 
profit the bot has generated, monitor your **free** or **available** BTC balance. 
It will indicate your actual gain. 

Just wait a little bit longer. As long as the bot is able to trade, everything is fine. 
Remember you've made an estimate on how low the coin value will ever drop - everything is fine, 
as long as the price is above. 

If the price does drop below your estimate minimum, well you can either sell everything and 
accept a rather big loss, or place additional buy order and back'em up some more - or just "sit 
it out" and hope for better times.. 
Lets just hope you've placed enough buy order to circumvent this situation in the first place.


## Getting started - Quick guide
1. Register your trading accounts. 
2. Generate exchange api credentials
3. Choose a market / coin pair you want to invest in and "farm".
4. Place your BUY and SELL orders
5. Download the latest fitting version from 
[releases](https://github.com/webklex/sstb/releases) or build your own by following the 
build instructions below 
6. Prepare the bot (See the [configuration](#configuration) for additional information)
7. Execute the bot
8. (optional) setup a service and have it always running 


## Getting started - Step by Step
### 1. Account registration
I recommend you use one account per traded market. 
This has several benefits:

- You can't have more then 200 open orders (at least on Binance. I could not find any 
information on it regarding Poloniex). A bot instance has easily up to 100 or 150 open orders,
depending on your personal preference.

- It is really easy to track the performance. The available balance is always the current gain.

- Each account has a rate limit - a limited amount of request per second the bot can perform. If 
the rate limit gets reached, the exchange won't accept any requests and orders might not get placed.

#### 1.1 Binance
Use this [link](https://www.binance.com/en/register?ref=KLK9HBCF) or the code `KLK9HBCF` if 
you sign up to get 10% of all payed spot trade fees payed back by Binance.

#### 1.2 Poloniex
Use this [link](https://poloniex.com/signup?c=4EJJK4JR) or the code `4EJJK4JR` if you sign 
up. This will support the future development of this bot.

### 2. API Keys
You'll need to create api key pairs. This will allow the bot to place orders and is required 
to monitor all open orders. 
If you plan to run the bot in the cloud, limit the access to that ip. 
Also make sure to **disable** or **remove** the right to withdraw funds (if possible). Only the 
ability to do so called "spot" trading is required.

#### 2.1 Binance
Go to the [api management](https://www.binance.com/en/my/settings/api-management) section and 
create a new key pair. You'll need to enable two factor authentication in order to do so.

#### 2.2 Poloniex
Go to [api keys](https://poloniex.com/apiKeys) and create a new key pair.

### 3. Choose a market / coin pair
I assume you have x amount of BTC deposited in the newly created account. If not you need to 
deposit the amount you want to invest (see [How much should I at least invest?](#q-how-much-should-i-at-least-invest) for 
additional information).

Now take a look at all available markets which can be traded with BTC and look for a pair which
fits your needs (see [Which market is the best?](#q-which-market-is-the-best) for additional information).

### 4. Place all necessary orders
Now you have to make two little prophecies:
- Your highest expected price
- Your lowest expected price

#### 4.1 Choose a trade volume
The chosen trade volume has to be above the allowed minimum. This number can vary depending on
the exchange and the market you want to use. However it is usually `0.0001 BTC` which I like to
increase a bit to be save. Instead I tend to go with a minimum of  `0.00010100 BTC` or 10100 
Satoshi.

If you increase the trade volume, you will also increase the gain - however you also have to 
invest more. For a first start I recommend you go with the minimum plus a little extra like I did.

#### 4.2 Choose a step size
The step size will ultimately determine the trade speed. In my own tests it turned out that the 
lowest step size had the highest profit, but it also requires a bigger investment.

Let me explain it with an example:
If you buy 10 **coinA** for 0.1 **coinB** each (a total of 1 **coinB**) and sell all 10 **coinA** 
for a total of 1.1 **coinB** (0.11 **coinB** per **coinA**), then your step size is 0.01. The step size is the amount that gets added to the 
single unit price - the price you pay for one **coinA**.

Here is another real world example:
```
Market:        DOGE/BTC
Current price: 0.00000050 BTC
Highest price: 0.00000070 BTC
Lowest price:  0.00000020 BTC
Step size:     0.00000001 BTC
Volume:        0.00010100 BTC
```
In order to prepare the bot for the above scenario, you'll need to buy enough DOGE to place the 
following sell orders, but still have enough BTC to also place the buy orders:

| Side   | Price      | Volume  | Total (BTC) |
| :----- | :--------- | :------ | :---------- |
| SELL   | 0.00000070 | 144     | 0.00010080  |
| ...    | ...        | ...     | ...         |
| SELL   | 0.00000053 | 190     | 0.00010070  |
| SELL   | 0.00000052 | 194     | 0.00010088  |
| SELL   | 0.00000051 | 198     | 0.00010098  |
| EMPTY  | EMPTY      | EMPTY   | EMPTY       |
| BUY    | 0.00000049 | 206     | 0.00010094  |
| BUY    | 0.00000048 | 210     | 0.00010080  |
| BUY    | 0.00000047 | 215     | 0.00010105  |
| ...    | ...        | ...     | ...         |
| BUY    | 0.00000020 | 505     | 0.00010100  |

The `EMPTY` market lines can be ignored. No order should be placed there. If the `SELL` order
for `0.00000051 BTC` gets fulfilled, the bot will automatically create a `BUY` order for 
`0.00000050 BTC` with a total of the defined trade volume. So in this case the following order 
will be created:

| Side   | Price      | Volume  | Total (BTC) |
| :----- | :--------- | :------ | :---------- |
| BUY    | 0.00000050 | 202     | 0.00010100  |

If this buy order gets fulfilled, the following order will be created:

| Side   | Price      | Volume  | Total (BTC) |
| :----- | :--------- | :------ | :---------- |
| SELL   | 0.00000051 | 202     | 0.00010302  |

Congratulation: you've just gained your first profit of `0.00000202 BTC`.

Now this does ignore the trade fees of 0.1 - 0.15%, which won't be ignored by the bot. But you get
the point I'm trying to make?

The bot actually has a tiny little feature which will automatically sell coin decimals. DOGE for 
example can only be traded in whole numbers (on Binance). However the bot gathers these leftover 
decimals and add them to the next order. They will appear as "LEND" or "GAVE" log entries. 

If you don't have the funds to create all of these orders, you could increase the step size. If
you use `0.00000002` instead of `0.00000001` you only have to place halve as many orders.

Be aware that the minimal step size might be limited by the exchange. It is very likely that other
markets might only allow a step size of 0.01.

### 5. Download the latest release
You can download precompiled builds from the [releases](https://github.com/webklex/sstb/releases) 
page. If you do so, please compare the hash sum of the extracted binary to make sure now one has changed it.

If you work under a debian like os, I recommend to place it under `/opt/sstb/sstb`

#### 5.1 Build from source
I highly recommend you build the bot yourself. This is fairly easy if you've used "go" before. 
If you haven't - well you'll need to "google" it :)

If you have go and all dependencies installed, you can either execute the `build.sh` or simply 
call `go build -o sstb -ldflags "-s -w"` inside the repository directory to build the latest
version.

Additional information and an example can be found in the [Build](#build) section.

### 6. Prepare and configure the bot
The bot gets its configuration from a bunch of `.json` files. By default they are located in a
folder called `config` and placed inside the directory the binary got executed from.

If you work under a debian like os, I recommend to place them under `/opt/sstb/config`.

For the following examples I'll make these assumptions:

- You want to use a Poloniex a account 
- You want to use a Binance a account
- You want to get notified by your mattermost server
- All your job files will be located under `/opt/sstb/config/jobs`
- You want to trade `DOGE/BTC` on both exchange platforms
- The log file is located under `/opt/sstb/logs/log.log`

#### 6.1 App
The main config file is the app config, usually located under `config/app.json`. This file
contains the information for all providers / exchanges (api key pairs) and a list of all
notifiers such as slack or mattermost.

```json
{
  "timezone": "UTC",
  "job-dir": "/opt/sstb/config/jobs",
  "provider": [
    {
      "name": "my-poloniex-acc",
      "exchange": "poloniex",
      "key": "API_KEY",
      "secret": "API_SECRET"
    },
    {
      "name": "my-binance-acc",
      "exchange": "binance",
      "key": "API_KEY",
      "secret": "API_SECRET"
    }
  ],
  "notifier": [
    {
      "driver": "mattermost",
      "name": "my-mattermost-server",
      "endpoint": "https://some.host.com",
      "token": "YOUR_SECRET_TOKEN",
      "channel": "CHANNEL_ID",
      "username": "USERNAME"
    }
  ]
}
```

You can use anything as name for the notifier or provider as long as they are unique. 
They will be used by your jobs to reference the used notifier or provider.

If you don't want to receive any notifications, just ignore the notifier property.

Additional information and an example can be found in the [Configuration](#configuration) section.

#### 6.2 Jobs
Now that we have configured the base app, we can continue and create the actual trade / monitoring
jobs.

Create a new file `/opt/sstb/config/jobs/first-binance-bot.json` with a content like this:

```json
{
  "provider": "my-binance-acc",
  "symbol": "DOGEBTC",
  "primary": "BTC",
  "volume":"0.00010200",
  "fee": "0.1",
  "step": "0.00000001",
  "enabled": true,
  "alerts": {
    "buy": true,
    "sell": true,
    "idle": 300,
    "summary": true
  },
  "notifier": ["my-mattermost-server"]
}
```

This job will trade `DOGE/BTC` on your defined Binance account with a step size of 
`0.00000001 BTC`, a volume of `0.00010200 BTC` and will send a notification via mattermost for 
all possible alerts (sell, buy, daily summary and an idle alert if nor order has been placed 
for over 300 minutes).


Create an other file `/opt/sstb/config/jobs/first-poloniex-bot.json` with a content like this:

```json
{
  "provider": "my-poloniex-acc",
  "symbol": "BTC_DOGE",
  "primary": "BTC",
  "volume":"0.00010200",
  "fee": "0.15",
  "buy-step": "0.00000001",
  "sell-step": "0.00000002",
  "enabled": true
}
```

This job will trade `DOGE/BTC` on your defined Poloniex account with a step size of 
`0.00000001 BTC` for any buy orders, a steps size of `0.00000002 BTC` for any sell orders, with 
a volume of `0.00010200 BTC` and wont send any notifications.

Example 1:

| Side   | Price      | Volume  | Total (BTC) |
| :----- | :--------- | :------ | :---------- |
| BUY    | 0.00000050 | 202     | 0.00010100  |

If this buy order gets fulfilled, the following order will be created:

| Side   | Price      | Volume  | Total (BTC) |
| :----- | :--------- | :------ | :---------- |
| SELL   | 0.00000052 | 202     | 0.00010504  |

Example 2:

| Side   | Price      | Volume  | Total (BTC) |
| :----- | :--------- | :------ | :---------- |
| SELL   | 0.00000050 | 210     | 0.00010250  |

If this buy order gets fulfilled, the following order will be created:

| Side   | Price      | Volume  | Total (BTC) |
| :----- | :--------- | :------ | :---------- |
| BUY    | 0.00000049 | 206     | 0.00010094  |

Please note that the symbol between poloniex and binance differs. You have to use the correct one.
Poloniex always has a `_` between the two asset pairs. Binance doesn't.

Additional information and two other examples can be found in the [Job](#job) section.

#### 6.3 Logging
In order to customize and control the logging output, you'll need to create a new file located 
under `/opt/sstb/config/log.json` with the following content:

```json
{
	"silent": false,
	"stdout": false,
	"debug": false,
	"file": "/opt/sstb/logs/log.log",
	"timestamp": true
}
```

If you don't want to log the timestamp or want to use a different log file, just update the 
correlating attribute.

Additional information and an example can be found in the [Logging](#logging) section.

### 7. Execute the bot
Now lets see if you've configured everything correct.
You can start the bot by calling the binary directly and if you use the same configuration as 
above, then you dont have to provide any additional parameters. If you did change some of
the values, make sure to checkout the [available arguments](#arguments) in order to load all 
required config files.

```
cd /opt/sstb
./sstb
```

You will see an output like this:
```
2021/02/11 21:45:53 [info] Config file loaded successfully
2021/02/11 21:45:53 [Loaded 2 jobs]
2021/02/11 21:45:53 [MY-BINANCE-ACC ORDER REGISTERED: 41569079337]
2021/02/11 21:45:53 [MY-BINANCE-ACC ORDER REGISTERED: 41085630270]
2021/02/11 21:45:53 [MY-BINANCE-ACC ORDER REGISTERED: 40946923116]
...
2021/02/11 21:45:53 [Subscribing to MY-BINANCE-ACC account update events..]
```

### 7.1 (optional) Service setup
If you plan to let the bot run for a longer period, I recommend you get yourself a small vps
and host the bot somewhere.

You'll need some basic linux knowledge to set it up. A step by step setup guide can be found 
[here](#service-setup-guide).

### 8. Verification / Monitoring
So what now? You'll have to monitor your bot from time to time. Make sure it is still running, 
it isn't sold out (is still trading) and is generating the profit you want. This can all be 
done by visiting your exchange or taking a look at the logfile.


## Dependencies
These are all used go dependencies:
- [github.com/fatih/color](https://www.github.com/fatih/color)
- [github.com/kelseyhightower/envconfig](https://www.github.com/kelseyhightower/envconfig)
- [github.com/oleiade/reflections](https://www.github.com/oleiade/reflections)
- [github.com/adshao/go-binance/v2](https://www.github.com/adshao/go-binance)
- [github.com/mitchellh/mapstructure](https://www.github.com/mitchellh/mapstructure)
- [github.com/gorilla/websocket](https://www.github.com/gorilla/websocket)
- [github.com/pkg/errors](https://www.github.com/pkg/errors)
- [github.com/scorredoira/email](https://www.github.com/scorredoira/email)

## Build
Start by downloading the latest source code version:
```bash
git clone https://github.com/webklex/sstb.git
cd sstb
git checkout VERSION_NUMBER
```
You can now either build a packaged release..
```bash
./build.sh
```
..or just compile the binary for your current os:
```bash
go build -o sstb -ldflags "-s -w"
```

Finally do a quick test and check if the binary actually executes:
```bash
./sstb -h
```


## Service Setup
I strongly recommend to host the bot somewhere and have it setup as a service. However 
you'll need some basic linux knowledge to do so.

### VPS Provider
Here are some vps provider and their prices as of 2021-02-06 for the cheapest available 
version (the cheapest once are more then powerful enough). I personally can recommend 
Hetzner - I'm very happy with them :)
- **Strato** (~1.00€/month)
- **Aruba** (~2.79€/month)
- **Hetzner** (~2.89€/month) 
Use this [link](https://hetzner.cloud/?ref=D3W5L3vOEnKk) if you sign up to support future 
development.
- **DigitalOcean** (~5.00$/month)
Use this [link](https://m.do.co/c/053446f35524) to receive a 100$ welcome bonus by 
DigitalOcean.

### Service Setup Guide
1.) SSH into the new vm
```
ssh root@YOUR_VPS_IP
```

2.) Download the latest release to `/opt/sstb` or build it yourself
```
mkdir /opt/sstb
mkdir /opt/sstb/config
mkdir /opt/sstb/logs

cd /opt/sstb

wget https://github.com/webklex/sstb/releases/download/VERSION_NUMBER/sstb

chmod +x ./sstb

./sstb -h
```

```
cd /opt

git clone https://github.com/webklex/sstb.git
cd sstb
git checkout VERSION_NUMBER

go build -o sstb -ldflags "-s -w"

./sstb -h
```

3.) Place your configuration files under `/opt/sstb/config`

4.) Create a new service `nano /etc/systemd/system/sstb.service`
```
[Unit]
Description=Bot
After=multi-user.target 
After=syslog.target 
After=network-online.target

[Service]
Type=simple

User=www-data
Group=www-data

WorkingDirectory=/opt/sstb
ExecStart=/opt/sstb/sstb

Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

5.) Test and start the service
```
systemctl start sstb.service
systemctl status sstb.service
systemctl stop sstb.service
systemctl restart sstb.service
```


## Arguments
```bash
./sstb -h
```

| Option        | Value  | Default            | Description |
| :------------ | :----- | :----------------- | :---------- |
| -config       | string | ./config/app.json  | Application config file |
| -job-dir      | string | ./config/jobs      | Folder containing all job configuration files |
| -log-config   | string | ./config/log.json  | Log config file |
| -logtimestamp | bool   | true               | Prefix non-access logs with timestamp |
| -logtostdout  | bool   | false              | Log to stdout instead of stderr |
| -output-file  | string | ./logs/log_NOW.log | Log output file |
| -debug        | bool   | false              | Enable the debug mode |
| -silent       | bool   | true               | Disable logging and suppress any output |
| -timezone     | string | UTC                | Application time zone |
| -version      | bool   | false              | Show version and exit |


## Configuration
Example `config/app.json`:
```json
{
  "timezone": "UTC",
  "job-dir": "/path/to/config/jobs",
  "provider": [
    {
      "name": "some-custom-name-or-id",
      "exchange": "poloniex",
      "key": "API_KEY",
      "secret": "API_SECRET"
    },
    {
      "name": "other-custom-name-or-id",
      "exchange": "binance",
      "key": "API_KEY",
      "secret": "API_SECRET"
    }
  ],
  "notifier": [
    {
      "driver": "mattermost",
      "name": "some-name-or-id",
      "endpoint": "https://some.host.tld",
      "token": "YOUR_SECRET_TOKEN",
      "channel": "CHANNEL_ID",
      "username": "USERNAME"
    },
    {
      "driver": "slack",
      "name": "sl-example",
      "endpoint": "https://hooks.slack.com/services/YOUR_SECRET_TOKEN"
    },
    {
      "driver": "email",
      "name": "em-example",
      "email": {
        "host": "some.host.tld",
        "port": "1025",
        "username": "USERNAME",
        "password": "PASSWORD",
        "sender": "email@some.host.tld",
        "name": "SSTP",
        "subject": "Notification",
        "to": ["me@some.host.tld"],
        "cc": ["someone@some.host.tld"],
        "bcc": ["hidden@some.host.tld"]
      }
    }
  ]
}
```

**Attributes**

| Key      | Type       | Description   |
| :------- | :--------- | :------------ |
| timezone | string     | Timezone used (default: UTC) |
| job-dir  | string     | Directory containing all job configuration files (default: `config/jobs/`) |
| provider | []provider | An array containing all supported providers and their api keys |
| notifier | []notifier | An array containing all supported notifier such as slack, mattermost, etc |

### Provider

**Attributes**

| Key      | Type   | Description                               |
| :------- | :----- | :---------------------------------------- |
| name     | string | A unique name or id                       |
| exchange | string | The exchange id ("binance" or "poloniex") |
| key      | string | API key                                   |
| secret   | string | API secret                                |

### Notifier

**Attributes**

| Key      | Type   | Description |
| :------- | :----- |:------------|
| name     | string | A unique name or id |
| endpoint | string | Socket or Api endpoint |
| username | string | Username if required |
| token    | string | Password or token if required |
| channel  | string | Channel ID (mattermost) |
| driver   | string | The used driver (mattermost, slack or email) |

#### Slack
This bot can send notifications to Slack via an Incoming Webhook - you can set these up in the 
[Slack Apps](https://api.slack.com/apps) area.

#### Email
The generated email notification isn't pretty but it displays the same information as on slack
or mattermost.

If you dont want to send an email in `cc` or `bcc` just leave them empty or remove the attribute from
your configuration file. 

### Jobs
All jobs are placed inside the in `config/app.json` defined `job-dir`. The default location is
`config/jobs/`.

Example `config/jobs/first-job.json`:
```json
{
  "provider": "some-provider-name",
  "symbol": "DOGEBTC",
  "primary": "BTC",
  "volume":"0.00010100",
  "fee": "0.1",
  "step": "0.00000001",
  "enabled": true,
  "alerts": {
    "buy": true,
    "sell": true,
    "summary": true
  },
  "notifier": ["some-notifier-name"]
}
```

Example `config/jobs/second-job.json`:
```json
{
  "provider": "some-provider-name",
  "symbol": "BTC_DOGE",
  "primary": "BTC",
  "volume":"0.00010100",
  "fee": "0.15",
  "buy-step": "0.00000001",
  "sell-step": "0.00000002",
  "enabled": true,
  "alerts": {
    "buy": true,
    "sell": false,
    "summary": true
  },
  "notifier": ["some-notifier-name"]
}
```

**Attributes**

| Key            | Type     | Description   |
| :------------- | :------- | :------------ |
| provider       | string   | Name of a provider defined inside your `config/app.json` file |
| symbol         | string   | Symbol of the chosen market |
| primary        | string   | Primary coin - coin used to pay for a buy order |
| volume         | string   | Buy trade volume |
| fee            | string   | Trading fee in percent |
| step           | string   | Default trading step size |
| buy-step       | string   | Desired trading step size for placing buy orders |
| sell-step      | string   | Desired trading step size for placing sell orders |
| enabled        | bool     | Won't execute if set to `false` |
| notifier       | []string | An array of notifier names defined inside your `config/app.json` file |
| alerts.buy     | bool     | Send a notification if a buy order is created |
| alerts.sell    | bool     | Send a notification if a sell order is created |
| alerts.idle    | int      | Send an idle alert if no order has been placed for a given number of minutes |
| alerts.summary | bool     | Send a daily trading summary (every day at 10:00am UTC) |

### Logging
Example `config/log.json`:
```json
{
	"silent": false,
	"stdout": false,
	"debug": false,
	"file": "/path/to/logs/log.log",
	"timestamp": true
}
```

**Attributes**

| Key       | Type   | Description                             |
| :-------- | :----- | :-------------------------------------- |
| silent    | bool   | Disable logging and suppress any output |
| debug     | bool   | Enable the debug mode                   |
| stdout    | bool   | Log to stdout instead of stderr         |
| file      | string | Log output file                         |
| timestamp | bool   | Prefix non-access logs with timestamp   |


## Tips & Tricks
### Dynamic order volume
If you want to use a dynamic order volume, just place multiple orders per position. You can place
as many as you want (only limited by the exchange).

### Tested markets
I've used and tested the bot with the following markets with different but positive results:
- DOGE/BTC ($$) 
- GTO/BTC ($$) 
- IOST/BTC ($) 
- ANKR/BTC ($$$+) 
- CELR/BTC ($$$+) 
- AKRO/BTC ($$) 


## Known issues
#### Lots of client errors in the log file
This happens if the exchange is offline or has some technical issues. There is nothing you 
can do about it. Just hope they fix it fast...

#### A trade was not placed
This can happen if:
- The exchange has connection issues
- You have connection issues

This is solved easily by manually creating the missing orders.

 
## Support 
If you encounter any problems or if you find a bug, please don't hesitate to create a new 
[issue](https://github.com/webklex/sstb/issues). 
However please be aware that it might take some time to get an answer. 
Off topic, rude or abusive issues will be deleted without any notice. 
 
If you need **commercial** support, feel free to send me a mail at github@webklex.com.  

 
## Donate
If you like what I did and want to give something back, you are welcome to do so :)
However using one of the referral links below also really helps.
```
BTC: 13LMyXYDSVxoc2Rsc8mQCfbs5Fxnd2HiuG
```

**Referral links**

| Service      | Link                                             | Feature            |
| :----------- | :----------------------------------------------- | :----------------- |
| Binance      | https://www.binance.com/en/register?ref=KLK9HBCF | 10% fee payback    |
| Binance      | https://www.binance.com/en/register?ref=CIO46IDJ | trade fee donation |
| Poloniex     | https://poloniex.com/signup?c=4EJJK4JR           |                    |
| Hetzner      | https://hetzner.cloud/?ref=D3W5L3vOEnKk          |                    |
| DigitalOcean | https://m.do.co/c/053446f35524                   | 100$ welcome bonus |


## Change log
Please see [CHANGELOG][link-changelog] for more information what has changed recently.


## Security
If you discover any security related issues, please email github@webklex.com instead of 
using the issue tracker.


## Credits
- [Webklex][link-author]
- [All Contributors][link-contributors]


## License
The MIT License (MIT). Please see [License File][link-license] for more information.

[ico-version]: https://img.shields.io/github/v/tag/webklex/sstb?style=flat-square&label=version
[ico-license]: https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square
[ico-hits]: https://hits.webklex.com/svg/webklex/sstb
[ico-downloads]: https://img.shields.io/github/downloads/webklex/sstb/total?style=flat-square

[link-author]: https://github.com/webklex
[link-repo]: https://github.com/webklex/sstb
[link-contributors]: https://github.com/webklex/sstb/graphs/contributors
[link-license]: https://github.com/webklex/sstb/blob/master/LICENSE
[link-changelog]: https://github.com/webklex/sstb/blob/master/CHANGELOG.md
[link-hits]: https://hits.webklex.com