# Waktu Solat
CLI tool to retrieve malaysia prayer time.

### Demo
![](demo.gif)

### Help
```shell
NAME:
   waktu-solat - Retrieve prayer time

USAGE:
   waktu-solat [global options] command [command options] [arguments...]

COMMANDS:
   get       Retrieve prayer time
   zone      List all accepted zone
   set-zone  Set default zone id
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --db DB_FILE    path to DB_FILE (default: "<CACHE_PATH>/waktu-solat.db")
   --debug, -d     enable debug logs (default: false) [$WS_DEBUG]
   --help, -h      show help (default: false)
   --output value  output mode [cli, alfred] (default: "cli") [$WS_MODE]
```

### Source
- Info pull from https://www.e-solat.gov.my/
