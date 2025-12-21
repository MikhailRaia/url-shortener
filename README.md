# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Профилирование памяти (pprof)

- **Команда сравнения**:

```
go tool pprof -top -base profiles/base.pprof shortener.exe profiles/result.pprof
```

- **Вывод**:

```
Build ID: C:\Users\Mikhail Raya\AppData\Local\go-build\93\9311f3b8460e281aa0ea2c772d47afd5640e9db5c0f7c9bcf2f5f34ef9970084-d\shortener.exe2025-12-21 22:50:02.5558784 +0300 MSK
Type: inuse_space
Time: 2025-12-21 22:46:36 MSK
Showing nodes accounting for -1026kB, 50.00% of 2052kB total
      flat  flat%   sum%        cum   cum%
   -1026kB 50.00% 50.00%    -1026kB 50.00%  runtime.allocm
         0     0% 50.00%     -513kB 25.00%  runtime.mcall
         0     0% 50.00%     -513kB 25.00%  runtime.mstart
         0     0% 50.00%     -513kB 25.00%  runtime.mstart0
         0     0% 50.00%     -513kB 25.00%  runtime.mstart1
         0     0% 50.00%    -1026kB 50.00%  runtime.newm
         0     0% 50.00%     -513kB 25.00%  runtime.park_m
         0     0% 50.00%    -1026kB 50.00%  runtime.resetspinning
         0     0% 50.00%    -1026kB 50.00%  runtime.schedule
         0     0% 50.00%    -1026kB 50.00%  runtime.startm
         0     0% 50.00%    -1026kB 50.00%  runtime.wakep
```
