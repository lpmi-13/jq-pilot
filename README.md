# JQ Pilot

Inspired by the `jq` wizardry of csibar and thomas-franklin, I decided I needed a micromaterial to up my jq game, so I made this to generate learning exercises for myself.

## Local runs

To start this up and try some exercises, just run

```
npm start --prefix frontend
```

and you should be off to the races.

You'll see a prompt in the UI at `localhost:8000`, and if you hit `localhost:8000/question`, you'll get the source json back for the first problem.

You need to POST the data in the required format to `localhost:8000/answer`, and if you're right, you'll get a success message.

## Iterations

1. First MVP is just three hardcoded json problems that can be worked through in sequence.

2. Second form will be dynamic problems where you can continuously fire GET requests at the question endpoint, and get a different problem each time you correctly solve it. If you just fire requests at the GET endpoint without sending the required answer, you'll get the same prompt.

It could be this is achieved by some dynamic route handling gin magic, but it also could be that we run this in a compose stack that seeds the question on container startup and a correct POST triggers the container to kill itself by terminating the gin process.

I have no idea where we're going to end up putting the complexity, but it'll definitely be somewhere! :laughing:
