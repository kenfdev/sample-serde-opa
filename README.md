# Serialize/Deserialize PartialQueries

#### Testing this program

```
go run main.go
```

The Marshalled struct is as follows:

```json
{
  "queries": [
    [
      {
        "index": 0,
        "terms": {
          "type": "ref",
          "value": [
            { "type": "var", "value": "data" },
            { "type": "string", "value": "partial" },
            { "type": "string", "value": "authz" },
            { "type": "string", "value": "allow" }
          ]
        }
      }
    ]
  ],
  "modules": [
    {
      "package": {
        "path": [
          { "type": "var", "value": "data" },
          { "type": "string", "value": "partial" },
          { "type": "string", "value": "authz" }
        ]
      },
      "rules": [
        {
          "head": {
            "name": "allow",
            "value": { "type": "boolean", "value": true }
          },
          "body": [
            {
              "index": 0,
              "terms": [
                { "type": "ref", "value": [{ "type": "var", "value": "eq" }] },
                { "type": "string", "value": "alice" },
                {
                  "type": "ref",
                  "value": [
                    { "type": "var", "value": "input" },
                    { "type": "string", "value": "user" }
                  ]
                }
              ]
            }
          ]
        },
        {
          "default": true,
          "head": {
            "name": "allow",
            "value": { "type": "boolean", "value": false }
          },
          "body": [
            { "index": 0, "terms": { "type": "boolean", "value": true } }
          ]
        }
      ]
    }
  ]
}
```

Current status is that it **panics**.

```
panic: assertion failed
```
