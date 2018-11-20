# Bandit-server

Bandit-server is a [Multi-Armed Bandit](http://en.wikipedia.org/wiki/Multi-armed_bandit) api server which needs no configuration 

## Getting Started

1. Install bandit-server. ``go get github.com/recoilme/bandit-server``
2. Run ```bandit-server --port=3000 --debug=true```


## Hits

Hits - это количество показов объявлений. Для каждого объявления передаются:

- название объявления, строка, это рука многорукого бандита  (arm)
- количество, (cnt) - целое число

Запрос идет на url http://localhost:3000/hits/domainid42 

где domainid42 - это группа, по которой считатеся статистика. В нашем случае domainId.

Метод - POST

Но может быть любая строка.

Статистика передается в виде json массива. Заголовок: Content-Type: application/json


Пример запроса: 
```
curl -X POST --data '[{"arm":"ads 1","cnt":1},{"arm":"ads 2","cnt":1}]' -H "Content-Type: application/json" http://localhost:3000/hits/domainid42
```

На этот запрос сервер ответит:


HTTP/1.1 200 OK


Текст ответа:
ok

Пример "кривого", неправильного запроса:

```
curl -v -X POST --data '[{"arm":"ads 1","cnt":1},{"arm":"ads 2","cnt":"2"]' -H "Content-Type: application/json" http://localhost:3000/hits/domainid42
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
curl -X POST --data '[{"arm":"ads 2","cnt":1}]' -H "Content-Type: application/json" http://localhost:3000/rewards/domainid42
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


# License

Bandit-server is released under the [MIT License](http://www.opensource.org/licenses/MIT).
