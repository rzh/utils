# Concept
- _**Client**_: performance client, which has HammerTime, mongo-sim, or mongo-perf installed and we will drive traffic from all of them
- _**Server**_: mongod or mongos instance. There must be one mongod/mongos instance running. The server shall have iostat/pidstat installed
- _**Parser**_: to parse the output of client based on client type, the results will be send to dashboard backend once it is ready
- All the logs will save into a fold corresponding to the test run. User can specify what logs to be pulled from each server in the JSON conf files.

# What MC do?
- Log into all _**servers**_ via ssh. Figure out pid for mongod/mongos. Start monitoring with pidstat and iostat
- Start traffic generator from _**clients**_
- Wait for traffic generator 
- Stop monitor on servers. Save all monitorring log into local report folder
- Retrieve server log into the local report folder
- Retrieve client log into the local report folder
- Save traffic generator screen output to the local report folder
- Analyze traffic generator log based on client type
- Analyze server performance monitoring log
- Generate JSON for reporting
- Report to the dashboard backend (not yet)

# Assumptions
- All the clients have the identical setup, that is all the log file, binaries located at the same path. You can just use the same command to start all the client. Please note, although the tool support multiple client, current only the first one is used
- All the server to be monitored has exact one instance of mongod/mongos running.
- This is only to manage tasks and monitor server during the test. It makes no correctness check on setup and state fo the mongo cluster except make sure there is one instance of mongod/mongos running on each server. 
- The host running this tool has ssh access to all the clients/server
- If there is no client specified, it assumes run from local. 
- No support for local mongod/mongos testing for now. To be changed is deemed necessary.

# Environment Setup
## Server

In order to monitor the server, you need install sysstat, which will bring in pidstat and iostat.

## Client

Nothing is required for the client

# Installtion

Assume you already have working go environment setup, if not, please follow the instruction http://golang.org/doc/install
<pre>
# go get github.com/rzh/utils/go/mc
# go install github.com/rzh/utils/go/mc
# mc -run test1
</pre>

# Configure Files
Example: https://github.com/rzh/utils/blob/master/go/run/runner.json

More Details TBA

# Command Line Options
TBA

# Add new Parser
TBA
