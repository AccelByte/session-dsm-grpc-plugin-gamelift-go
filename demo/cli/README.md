# Session Service's Custom DSM Plugin gRPC Demo App

A CLI demo app to prepare required data and execute Custom DSM Plugin gRPC for AGS's Session Service.
Following diagram will explain how this CLI demo app works.
```mermaid
sequenceDiagram
    participant A as CLI Demo App
    participant I as AGS IAM
    participant P as AGS Session
    participant G as Grpc Plugin Server / Extend App
    
    A ->> I: user login
    I -->> A: auth token
    A ->> P: Create session template using custom grpc server target or extend app
    A ->> I: Create new test player
    A ->> P: Create game session
    P ->> G: Call CreateGameSession
    P -->> A: Returns session info
    A ->> P: Check game session for DS information
    G -->> P: Returns new DS information
    P -->> A: Returns session info with DS information    
    A ->> P: Delete session
    P ->> G: Call TerminateGameSession
    A ->> I: Delete test player
    A ->> P: Delete session template
```

## Prerequsites

- Docker
- make

## Usage

### Setup

The following environment variables are used by this CLI demo app.
```
export AB_BASE_URL='https://test.accelbyte.io'
export AB_CLIENT_ID='xxxxxxxxxx'
export AB_CLIENT_SECRET='xxxxxxxxxx'
export AB_NAMESPACE='namespace'

export GRPC_SERVER_URL='<gRPC server url accessible by internet>'
```
If these variables aren't provided, you'll need to supply the required values via command line arguments.

Also, you will need `Custom Session DSM Plugin gRPC` server already deployed and accessible. If you want to use your local development environment, you can use tunneling service like `ngrok` to tunnel your grpc server port so it can be accessed by AGS.
> Current AGS deployment does not support mTLS and authorization for custom grpc plugin. Make sure you disable mTls and authorization in your deployed Grpc server.


### Run

Run the demo test cli using makefile with provided .env file containing required variables

```bash
$ make run ENV_FILE_PATH=.env
```