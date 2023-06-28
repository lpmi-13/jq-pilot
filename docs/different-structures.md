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
      "PurchaseCode": 801,
      "PurchasePrice": 3.50
    },
    {
      "PurchaseID": "fdae32eafb294",
      "PurchaseCurrency": "TMT",
      "PurchaseItem": "boat",
      "PurchaseCode": 801,
      "PurchasePrice": 22.00
    },
    {
      "PurchaseID": "b07fd0d96cff4f",
      "PurchaseCurrency": "RUB",
      "PurchaseItem": "bus",
      "PurchaseCode": 3997,
      "PurchasePrice": 45.00
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

for a bit of object transformations, we can have something similar to the tags object returned from AWS APIs, like turning:

```
[
  {
    "label": "house",
    "value": "lloyds pharmacy"
  },
  {
    "label": "house_number",
    "value": "105"
  },
  {
    "label": "road",
    "value": "church road"
  },
  {
    "label": "postcode",
    "value": "bn3 2af"
  }
]
```

into...

```
{
  "house": "lloyds pharmacy",
  "house_number": "105",
  "road": "church road",
  "postcode": "bn3 2af"
}
```

via some old `jq 'map({ (.label): .value }) | add '` magic (to be fair, that might be the only thing we practice, since it's too cool to not practice...maybe also do something with both `to_entries`, `with_entries`, and `from_entries`, cause I have no idea what they do)
