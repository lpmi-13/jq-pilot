# JQ Pilot

Inspired by the `jq` wizardry of @csibar and @thomas-franklin, I decided I needed a micromaterial to up my jq game, so I made this to generate learning exercises for myself.

live at https://jkew.party (and also https://jayq.party, since I couldn't decide which name I like more).

## The exercises

based on: https://gist.github.com/olih/f7437fb6962fb3ee9fe95bda8d2c8fa4

-   grab the raw value of a key
-   filter for one specific key
-   delete one specific key
-   keep one specific key
-   get a deeply nested value
-   get all values for repeated nested key
-   get all unique values for repeated nested key
-   get number of keys
-   get length of array
-   get range of values from array (eg, [2:])
-   filter array for all integer values above/below a value
-   get the type of array item
-   get min and max values of an array
-   convert array of verbose label/value objects to simple key/value objects (from_entries)
-   convert a simple key/value pair object into a verbose label/value array (to_entries)
-   make all string values lowercase
-   add a new property to every object
-   group objects by a common key (group_by)

> If you need some help with specific excercises, you can look at this [cheat sheet](docs/cheat-sheet.md)

## Local runs

There are two ways to run this locally, depending on whether you have node and go installed, or just want to use docker.

### Node and Go installed

To start this up and try some exercises, just run

```
go mod download
npm i --prefix frontend
npm start --prefix frontend
```

and you should be off to the races.

### The Docker Way

Be sure to pass in `--build-arg ENV=local` (or any value that's not "production") to use `localhost` as the domain.

```
docker build -t jq-pilot --build-arg ENV=local .
docker run -it --rm -p 8000:8000 jq-pilot
```

### Doing the exercises

The recommended way to interact with this is to use `curl` in a terminal, but feel free to interact with the endpoint via the tool of your choice!

You'll see a prompt in the UI at `localhost:8000`, and if you hit `localhost:8000/question`, you'll get the source json back for the first problem.

You need to POST the data in the required format to `localhost:8000/answer`, and if you're right, you'll get a success message.

Helpful tip: if you want to pipe something into curl, you can do the following:

```
$ curl localhost:8000/question | jq DO_STUFF_HERE | curl -d @- localhost:8000/answer
```

I added a `/prompt` endpoint to the webserver just in case you want to try and interact with this only via the command line, though it doesn't necessarily tell you what was wrong about a wrong answer, only that it wasn't correct, so this endpoint might have limited value.

## Development

If you just wanna hack on the gin webserver, you can run go with `nodemon` to get live reloading on code changes, which is hashtag delicious!

```
nodemon --exec go run main.go --signal SIGTERM
```

## The commands we're practicing

For a refresher on some of the useful `jq` commands, see [this doc](docs/useful-commands.md)

## Different types of data structures

For an overview of the different json data schemas, see [this doc](docs/different-structures.md)

## Super Secret Section

...and if you need a cheat sheet, then go ahead and peep [this](docs/cheat-sheet.md)
