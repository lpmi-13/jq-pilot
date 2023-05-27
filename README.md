# JQ Pilot

Inspired by the `jq` wizardry of @csibar and @thomas-franklin, I decided I needed a micromaterial to up my jq game, so I made this to generate learning exercises for myself.

live at https://jkew.party

> but also, since https://jkew.party is currently hosted on railway, their free tier runs out after 21 days, so if it's after the 21st of the month, check it out at https://jq-pilot.fly.dev

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
-   get range of values from array (eg, [2:4])
-   filter array for all integer values above/below a value
-   get the type of array item
-   get min and max values of an array

## Local runs

To start this up and try some exercises, just run

```
go mod download
npm i --prefix frontend
npm start --prefix frontend
```

and you should be off to the races.

The recommend way to interact with this is to use `curl` in a terminal, but feel free to interact with the endpoint via the tool of your choice!

You'll see a prompt in the UI at `localhost:8000`, and if you hit `localhost:8000/question`, you'll get the source json back for the first problem.

You need to POST the data in the required format to `localhost:8000/answer`, and if you're right, you'll get a success message.

Helpful tip: if you want to pipe something into curl, you can do the following:

```
$ curl localhost:8000/question | jq DO_STUFF_HERE | curl -X POST -d @- localhost:8000/answer
```

## Development

If you just wanna hack on the gin webserver, you can run go with `nodemon` to get live reloading on code changes, which is hashtag delicious!

> nodemon --exec go run main.go --signal SIGTERM

## Commands to practice (taken from https://cameronnokes.com/blog/jq-cheatsheet/)

-   get a named property

```
echo '{"id": 1, "name": "Cam"}' | jq '.id'``
# 1
```

or

```
echo '{"nested": {"a": {"b": 42}}}' | jq '.nested.a.b'
# 42
```

-   get an array element's properties

```
echo '[{"id": 1, "name": "Mario"}, {"id": 2, "name": "Luigi"}]' | jq '.[1].name'
# Luigi
```

-   select specific key/value pair

```
echo '{"id": 123, "name": "Cam", "location": "Earth"} | jq '{"location"}'
# {"location": "Earth"}
```

-   filter out specific key

```
echo '{"id": 123, "name": "Adam", "location": "Mars"} | jq 'del(.id)'
# {"name": "Adam", "location": "Mars"}
```

-   get an array element by index

```

echo '[0, 1, 1, 2, 3, 5, 8]' | jq '.[3]'

# 2

```

-   slice an array

```

echo '["a", "b", "c", "d"]' | jq '.[1:3]'

# ["b", "c"]

```

-   creating a new object

```

echo '{ "a": 1, "b": 2 }' | jq '{ a: .b, b: .a }'

# { "a": 2, "b": 1 }

```

-   get an objects keys as an array

```

echo '{ "a": 1, "b": 2 }' | jq 'keys'

# [a, b]

```

-   get the length of an array

```

echo '[0, 1, 1, 2, 3, 5, 8]' | jq 'length'

# 7

```

-   get the number of keys

```

echo '{"a": 1, "b": 2}' | jq 'length'

# 2

```

-   condense a nested array into a flat array

```

echo '[1, 2, [3, 4]]' | jq 'flatten'

# [1, 2, 3, 4]

```

-   get unique values in an array

```

echo '[1, 2, 2, 3]' | jq 'unique'

# [1, 2, 3]

```

## Different types of structures

This is for simple things like named property selection, filtering out certain keys, and selecting certain keys. We'll probably also use it for nested values.

```
{
  "activities": {
    "friday": "fishing",
    "monday": "reading",
    "saturday": "knitting",
    "sunday": "hiking",
    "thursday": "baking",
    "tuesday": "puzzles",
    "wednesday": "tennis"
  },
  "age": 13,
  "favoriteColors": [
    "brown", "black", "indigo"
  ],
  "id": 22828385,
  "location": "London",
  "name": "Pat"
}
```

We also want some choices that involve doing things with arrays so we can pull out all values for a specific property into an array, and optionally get all the unique ones. We can also count the number of occurrences of things at various levels of nesting.

```
{
  "purchases": [
    {
      "PurchaseID": "2ac8048b10b1",
      "PurchaseCurrency": "KYD",
      "PurchaseItem": "plane",
      "PurchaseCode": 801
    },
    {
      "PurchaseID": "fdae32eafb294",
      "PurchaseCurrency": "TMT",
      "PurchaseItem": "boat",
      "PurchaseCode": 801
    },
    {
      "PurchaseID": "b07fd0d96cff4f",
      "PurchaseCurrency": "RUB",
      "PurchaseItem": "bus",
      "PurchaseCode": 3997
    }
  ]
}
```

For doing things like finding unique values and length of arrays, we'll
have a super simple data structure that's just lottery picks

```
{
  "lotteryPicks": [
    {
      "Person": "Charity",
      "Numbers": [
        2, 1, 8, 0, 2
      ]
    },
    {
      "Person": "Bennett",
      "Numbers": [
        1, 6, 3, 7, 6
      ]
    }
  ]
}
```

For filtering by min/max values, we'll have some student grading data. This will probably be the complex one, where we practice object constructors, and filtering by
max/min value.

[This](https://earthly.dev/blog/jq-select/) is an EXCELLENT resource for some of the more complex commands

```
{
    "students": [
        "Joe": {
          "grades": {
            "math": [
                82, 90, 74, 88, 93, 80
            ],
            "art" : [
                85, 95, 72, 56, 80, 77
            ],
            "history": [
                67, 77, 68, 81, 74, 70
            ]
          }
        },
       "Susan": {
          "grades": {
            "math": [
                82, 90, 74, 88, 93, 80
            ],
            "art" : [
                85, 95, 72, 56, 80, 77
            ],
            "history": [
                67, 77, 68, 81, 74, 70
            ]
          }
        },
        "Cameron": {
          "grades": {
            "math": [
                82, 90, 74, 88, 93, 80
            ],
            "art" : [
                85, 95, 72, 56, 80, 77
            ],
            "history": [
                67, 77, 68, 81, 74, 70
            ]
          }
        },
        "Emily": {
          "grades": {
            "math": [
                82, 90, 74, 88, 93, 80
            ],
            "art" : [
                85, 95, 72, 56, 80, 77
            ],
            "history": [
                67, 77, 68, 81, 74, 70
            ]
          }
       }
    ]
}
```

## Iterations

1. First MVP is just three hardcoded json problems that can be worked through in sequence.

> this is done.

2. Second form will be dynamic problems where you can continuously fire GET requests at the question endpoint, and get a different problem each time you correctly solve it. If you just fire requests at the GET endpoint without sending the required answer, you'll get the same prompt.

> this is also done.

3. Third step is to start adding different types of activities to cover all the common use cases for JQ
