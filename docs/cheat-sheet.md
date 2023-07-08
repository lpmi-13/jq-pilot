# Cheat sheet for JQ Pilot exercises

> NOTE: pay attention to which ones expect to be returned inside an array/object, and which ones don't

-   delete a key

```
jq 'del(.key)'
```

-   select student who is the best at a certain subject

```
jq '([.students[] | {name: .name, score: .grades.results.art | add}] | max_by(.score).name) as $name | .students[] | select(.name == $name)'
```

-   find the number of array elements (if `key` is an array)

```
jq '.key | length'
```

-   get the highest grade for a student in a particular subject

```
jq '.students[] | select(.name == "STUDENT_NAME") | .grades.results.history | max'
```

-   get the top scores for each student in each subject

```
jq '.students[] |
```

-   pick one lottery contestant to be the winner

```
jq '.lotteryPicks[] | select(.person == "WINNER_NAME") | .winner = true'
```

-   find all the purchase currencies

```
jq '[.purchases[] | .PurchaseCurrency]'
```

-   find all purchases with a price above X

```
jq '[.purchases[] | select(.PurchasePrice > X)]'
```

-   get all unique lottery picks

```
jq '[.lotteryPicks[] | .numbers] | flatten | unique'
```

-   find the first N lottery picks for each person

```
jq '[.lotteryPicks[] | .numbers = .numbers[:N]]'
```

-   find the last N lottery picks for each person

```
jq '[.lotteryPicks[] | .numbers = .numbers[-N:]]'
```

-   find out how frequently each lottery number was picked

```
jq '[.lotteryPicks[] | .numbers] | flatten | map(tostring) | group_by(.) | map({(.[0]): length}) | add'

-   get the highest grade for each student in each subject

```

jq '[.students[] | {name: .name, grades: { results: { art: .grades.results.art | max, history: .grades.results.history | max, math: .grades.results.math | max}}}]'

```


-   turn a dictionary of key/value pairs into an array of labels/values ordered by label

```

jq 'to_entries | map({key: .key, label: .value}) | sort_by(.label)'

```

-   turn an array of labels/values into a key/value dictionary

jq 'map({ (.label): .value }) | add '
```
