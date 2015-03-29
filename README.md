## MOMA

1. create your config env folder
```
mkdir config
cd config
mkdir dev
```
2. create your app.json
```
{
    "port": ":9000",
    "redis" : {
        "host": "redisserver",
        "port": ":6379"
    },
    "personaurl": "http://127.0.0.1:9000"
}
```
3. create your auth.json
```
{
    "normals":{
        "test@po.com":true
    },
    "admins":{
        "test@po.com":true
    }
}
```
run package.sh to package evrtg
