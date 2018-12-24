# Bandit-server

Bandit-server is a [Multi-Armed Bandit](http://en.wikipedia.org/wiki/Multi-armed_bandit) api server which needs no configuration 

## Getting Started

1. Install bandit-server. ``go get github.com/recoilme/bandit-server``
2. Run ```bandit-server --port=3000 --debug=true```

## Routers

```
GET    /                         --> main.ok (3 handlers)    - for check status
GET    /stats/:group/:count      --> main.stats (3 handlers) - for get stats by count arms
POST   /stats/:group/:count      --> main.stats (3 handlers) - for get stats by arms
POST   /write/:param/:group      --> main.write (3 handlers) - for write hits & rewards
```

## Hits

Hits - это количество показов объявлений. Для каждого объявления передаются:

- название объявления, строка, это рука многорукого бандита  (arm)
- количество, (cnt) - целое число

Запрос идет на url http://localhost:3000/write/hits/domainid42 

где domainid42 - это группа, по которой считается статистика. В нашем случае domainId. Но может быть любая строка.

Метод - POST


Статистика передается в виде json массива. Заголовок: Content-Type: application/json


Пример запроса: 
```
curl -X POST --data '[{"arm":"ads 1","cnt":1},{"arm":"ads 2","cnt":1}]' -H "Content-Type: application/json" http://localhost:3000/write/hits/domainid42
```

На этот запрос сервер ответит:


HTTP/1.1 200 OK


Текст ответа:
ok

Пример "кривого", неправильного запроса:

```
curl -v -X POST --data '[{"arm":"ads 1","cnt":1},{"arm":"ads 2","cnt":"2"]' -H "Content-Type: application/json" http://localhost:3000/write/hits/domainid42
```
Ответ с ошибкой:
```
HTTP/1.1 422 Unprocessable Entity
Текст ответа:
{"error":"invalid character '\"' after array element"}
```
422 - это код ошибки.

## Rewards

Rewards - это награда за клик по объявлению.
Формат точно такой же. Единственное отличие - в url вместо hits передается rewards

Пример:
```
curl -X POST --data '[{"arm":"ads 2","cnt":1}]' -H "Content-Type: application/json" http://localhost:3000/write/rewards/domainid42
```

## Stats

Статистика запрашивается в следуещем формате:

curl -X GET http://localhost:3000/stats/domainid42/2

где domainid42 - это наша группа, 2 - Это количество рук, которые надо вернуть

Пример ответа:

```
[{"arm":"var2","hit":1,"rew":0,"score":1.9727697022487511},{"arm":"ads 2","hit":3,"rew":1,"score":1.4723124519757878}]
```
Это массив, он отсортирован по параметру score. 

Также можно запросить статистику по переданным "рукам", при помощи Post запроса с массивом "рук" 


Пример:

```
curl -X POST --data '[{"arm":"ads 2"},{"arm":"1"}]' -H "Content-Type: application/json" http://localhost:3000/stats/domainid42/3

Ответ:
[{"arm":"1","hit":0,"rew":0,"score":100},{"arm":"ads 2","hit":3,"rew":2,"score":1.5224751688711062}]
```

## Debug

Для запуска в режиме отладки необходимо передать флаг debug=true

Пример отладчика:

```
./bandit-server --debug=true
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /                         --> main.ok (3 handlers)
[GIN-debug] GET    /stats/:group/:count      --> main.stats (3 handlers)
[GIN-debug] POST   /:param/:group            --> main.write (3 handlers)
[{ads 1 1} {ads 2 1}]
[GIN] 2018/11/20 - 19:06:34 | 200 |    2.767846ms |             ::1 | POST     /hits/domainid42
[{ads 1 1} {ads 2 1}]
[GIN] 2018/11/20 - 19:06:56 | 200 |     386.706µs |             ::1 | POST     /hits/domainid42
[{ads 2 1}]
[GIN] 2018/11/20 - 19:07:12 | 200 |    2.396616ms |             ::1 | POST     /rewards/domainid42
[GIN] 2018/11/20 - 19:07:21 | 200 |     710.533µs |             ::1 | GET      /stats/domainid42/2
```


## Backup


Надо выполнить запрос:
``` 
curl -X GET http://localhost:3000/backup/backup
```
последний параметр - backup - директория.

Ответ - 200 ok

# License

Bandit-server is released under the [MIT License](http://www.opensource.org/licenses/MIT).
