# Serialize/Deserialize PartialQueries

#### Testing this program

```
go run main.go
```

The Marshalled struct is as follows:

```json
{
  "Query": "data.partial.authz.allow",
  "Support": "package partial.authz\n\nallow = true { \"alice\" = input.user }\ndefault allow = false"
}
```

The Eval succeeds both before the serialization and after the deserialization

```
1st ResultSet: [{[true] map[]}]
2nd ResultSet: [{[true] map[]}]
```
