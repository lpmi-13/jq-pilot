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
