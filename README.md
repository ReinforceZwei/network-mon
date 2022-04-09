# network-mon

A small program to monitor network connectivity and restart network/reboot router through ssh.

# Why

Just because my home network is going to disconnect randomly without any reason, about once per two weeks. Simply restarting the network service on my router instantly restore the network connectivity. So I decided to _automate_ this process before I solve the underlying problem.

# How does it work

This program was designed to run on your server within the LAN.

It will continuously pinging Internet hosts to monitor the network connectivity. If the pings keep failing, it will attempt to restart network service on the router and test again. If it doesn't work, it will ultimately attempt to reboot the router. After the network is resumed, a notification will be sent through Discord Webhook. 

Unfortunately it cannot send you a Discord alert when the Internet is down :P

# Config

**Please safely keep the config file. Exposing your config file with your router login password could cause your router being hacked.**

The program will create a default config file `netmonconfig.json` at first time running. Please review and edit the config file before using the program.

Your router must support ssh protocol and provide command for restarting network/rebooting.
```json
{
    "Router": "192.168.1.1", // IP address of the router
    "SshUser": "admin",      // ssh user
    "SshPassword": "admin",  // ssh password
    "TestTarget": [          // Which target to ping
        "1.1.1.1",           // First address will always be used
        "8.8.8.8",           // Other address will be used as fallback when ping to first address was failed
        "8.8.4.4"
    ],
    "TestIntervalSecond": 300,    // How long to wait after successful ping
    "FailingIntervalSecond": 10,  // How long to wait after failure ping
    "MaxFailAttempt": 10,         // How many ping attempts before restarting network
    "NetworkRestartCommand": "service restart_wan", // Command to restart command through ssh
    "RestartAttempt": 3,          // How many restart attempts if ping still fail after restart
    "RestartMaxWaitTime": 40,     // How long to wait after issuing restart command
    "RouterRebootCommand": "reboot", // Command to reboot through ssh
    "RebootAttempt": 1,           // How many reboot attempts
    "RebootMaxWaitTime": 120,     // How long to wait after issuing reboot command
    "Notify": {                   // Discord Webhook for receiving notification after network went down and resumed
        "Url": "",
        "MessagePayload": "{\"content\":\"%s\"}" // See https://discord.com/developers/docs/resources/webhook#execute-webhook-jsonform-params
    }
}
```
