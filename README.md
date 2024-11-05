<h1 align="center">convoC2</h1>

## Table of Contents
- [Introduction](#introduction)
- [Demo](#demo)
- [Download](#download)
- [Usage](#usage)
- [Requirements](#requirements)
- [Inspiration](#inspiration)
- [Contribute](#contribute)

## Introduction
Command and Control infrastructure that allows Red Teamers to execute system commands on compromised hosts through Microsoft Teams.  
It infiltrates data into hidden span tags in Microsoft Teams messages and exfiltrates data by embedding command outputs in Adaptive Cards image URLs, triggering out-of-bound requests to a C2 server.  
The lack of direct communication between the victim and the attacker, combined with the fact that the victim only sends http requests to Microsoft servers and antiviruses don't look into MS Teams log files, makes detection more difficult.  

![convoC2-Architecture](https://github.com/user-attachments/assets/d126a4cb-dc62-4a18-8b89-3501a4319d6e)

## Demo
The following video demonstrates the use of the server to control two compromised hosts: one running the new Teams on Windows 11 and the other running the old Teams on Windows 10. In the first case, the attacker is not in the same org as the victim.

https://github.com/user-attachments/assets/d9fcccc1-de73-49cd-aaa7-df53303dbc6a

**Note:** In the first case the victim already accepted the chat with the external attacker, but in a real scenario where the attacker starts the chat with the victim for the first time, the victim would need to confirm the chat with the attacker. This is not a problem: because messages are being cached in the log file anyways, the **commands are received and executed even if the victim has not visualized or accepted the chat yet**.

## Download
Download the latest version of the server:
```
#Amd64
wget https://github.com/cxnturi0n/convoC2/releases/download/v0.1.0-alpha/convoC2_server_amd64.tar.gz
tar -xzvf convoC2_server_amd64.tar.gz --one-top-level

#Arm64
wget https://github.com/cxnturi0n/convoC2/releases/download/v0.1.0-alpha/convoC2_server_arm64.tar.gz
tar -xzvf convoC2_server_arm64.tar.gz --one-top-level
```
Download the latest version of the agent:
```
wget https://github.com/cxnturi0n/convoC2/releases/download/v0.1.0-alpha/convoC2_agent.tar.gz
tar -xzvf convoC2_agent.tar.gz --one-top-level
```
## Usage
**Server**
```
root@convoC2-server-VPS:~# ./convoC2_server_amd64 -h
Usage of convoC2 server:
  -t, --msgTimeout  How much to wait for command output (default 30 s)
  -b, --bindIp      Bind IP address (default 0.0.0.0)
```
**Agent**
```
C:\Windows>convoC2_agent.exe -h
Usage of convoC2 agent:
  -v, --verbose   Verbose logging (default false)
  -s, --server    C2 server URL (i.e. http://10.11.12.13/)
  -t, --timeout   Teams log file polling timeout [s] (default 1)
  -w, --webhook   Teams Webhook POST URL
  -r, --regex     Regex to match command (default "<span[^>]*aria-label=\"([^\"]*)\"[^>]*></span>")
```

## Requirements
To get it working, you will need to set up a few things:
- [**Create Teams channel with Workflow Incoming Webhook**](#Create-Teams-channel-with-Workflow-Incoming-Webhook): this is the place where the adaptive cards containing the output will be received. <ins>**It is important to keep a browser window with this channel opened while using the server**</ins>, otherwise the server will not receive messages from the agents.

![WebhookChannel](https://github.com/user-attachments/assets/ea65daad-9274-4574-835b-107f468a1d6e)

- [**Fetch Ids and Auth Token**](#Fetch-Ids-And-Auth-Token): Teams initializes a chat with a POST to `https://teams.microsoft.com/api/chatsvc/emea/v1/threads` with the unique ids of the victim and the attacker in the body. In the response, the threadId will be returned in the path of the Location header url.
  The Bearer token of the same request is used to authenticate to `https://teams.microsoft.com/api/chatsvc/emea/v1/users/ME/conversations/<threadId>/messages`, which is the endpoint for sending messages. So we just need to grab these three things and the server will take care of the rest.
- Make sure you have a public facing host allowing inbound HTTP traffic on port 80.
- Teams needs to be running on the victim host, in the background is fine too.

After starting the server you can receive new agents and control them, using the data previously obtained for authentication. Check the demo out for a usage example.

### Create Teams channel with Workflow Incoming Webhook
First of all, you will need to create a Teams channel.  

![CreateTeams](https://github.com/user-attachments/assets/8f4da165-4c94-4685-b06c-a938cde8cf90)  

Right click on the three dots, then click on "Workflows".    

![CreateTeams_1](https://github.com/user-attachments/assets/50c0fe37-d81c-4fc6-b9ec-81a287cbcedd)  

In the search bar type "Webhook" and click on the "Post to a channel when a webhook request is received".  

![CreateTeams_5](https://github.com/user-attachments/assets/3f7d7221-cc84-457b-9057-610e60944c61)  

Continue with the default settings and finally copy the url.  

![Create_Teams_6](https://github.com/user-attachments/assets/45030146-3575-4588-9fac-ee2bb6079bbc)


### Fetch Ids and Auth Token
Start by looking for the victim.  

![GetOrgIds](https://github.com/user-attachments/assets/c75b47e3-5503-473b-8be6-a04d041e9422)  

After selecting the victim account, with the web proxy intercepting the requests, you can send a dummy message.  

![GetOrgIds_1](https://github.com/user-attachments/assets/76533c54-2c71-4d0d-94dc-5d638a33242c)  

Save the two ids present in the `api/chatsvc/emea/v1/threads` request then DROP the request. The auth token will be the Bearer token of the same request.  

![GetOrgIds_2](https://github.com/user-attachments/assets/b0f9d9bc-df2d-4527-8a87-301babd505b4)

## Inspiration
This infrastructure was inspired by the [Teams GIFShell research](https://medium.com/@bobbyrsec/gifshell-covert-attack-chain-and-c2-utilizing-microsoft-teams-gifs-1618c4e64ed7) made by [Bobbyrsec](https://www.linkedin.com/in/bobby-rauch/).  
Initially my purpose was to replicate the issue, but the solution appeared to be partially fixed. The research involved injecting commands into Base64-encoded GIFs, but these were no longer being displayed correctly: an icon of corrupted image would appear in the chat instead, resulting in the victim protentially starting to suspect.  
After testing various methods, I noticed that it was possible to **embed commands directly into messages rather than in GIFs or images**: sending multiple images is more likely to raise suspicion compared to sending simple messages, right?
Initially I considered embedding commands in tags like `<script>`, `<meta>`, or custom tags, but all were filtered out. Eventually, embedding commands in the `aria-label` attribute of `<span>` tags with `display:none` was successful: the victim would only see the message, but the hidden command would actually poison the Teams log file.  

![TeamsLogPoisoning](https://github.com/user-attachments/assets/b602ef7f-fabc-42a0-bd68-6b7d7f152a86)

The server TUI has been developed using the awesome Go [BubbleTea framework](https://github.com/charmbracelet/bubbletea).

## Contribute
If you find bugs or want to improve the project feel free to open a pull request and I will be glad to review and eventually merge your changes.
Short term todos are:
- Message AES encryption
- Keepalive to detect when agent is dead
- Powershell version of the Agent
