# Diplodocus

Diplodocus allows you to `tail -F` text files over http.

![Diplodocus](diplodocus.gif?raw=true)

## Fast start with docker

Expose files in `/var/log/whatever` on `192.168.0.7:8000`:

```
docker run -d -p 192.168.0.7:8000:8000 \
    -v /var/log/whatever:/var/log/diplodocus \
    bobrik/diplodocus
```

Now to follow `/var/log/whatever/mylog/mylog-2015-02-24.log`:

```
curl -v http://192.168.0.7:8000/mylog/mylog-2015-02-24.log
```

## Building

You will need [go compiler](http://golang.org/) installed.

```
mkdir diplodocus
cd diplodocus
export GOPATH=`pwd`
go get github.com/Topface/diplodocus/cmd/diplodocus-server
```

This will give you binary in `bin/diplodocus-server` that is ready to use.

## Running

```
./bin/diplodocus-server -listen 127.0.0.1:8000 -root /var/log/whatever
```

This will start http server on 127.0.0.1:8000. Any log in `/var/log/whatever`
can be monitored with command like this:

```
# monitor /var/log/whatever/example.com/access.log
curl -sN http://127.0.0.1:8000/example.com/access.log
```

Diplodocus will monitor for file updates, symlink changes
and whatever can happen to your logs to provide you with
constant stream of updates.

Hide it behind nginx or whatever proxy you like to manage access rights.
## Library

Diplodocus also provides a library for you to use, see
[server](cmd/diplodocus-server/main.go) code for example.

## Authors

* [Ian Babrou](https://github.com/bobrik)

## License

MIT
