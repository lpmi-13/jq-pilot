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
