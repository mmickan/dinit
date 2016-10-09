# Dinit


## Synopsis

    ./dinit [OPTIONS] -r CMD [OPTIONS..] [-r CMD [OPTIONS...]...]


## Description

Docker-init or dinit is a small init-like "daemon" (it is not a daemon) for use
within Docker containers.

Dinit will pass any environment variables through to the programs it is
starting. It will pass signals (SIGHUP, SIGTERM and SIGINT) through to the
children it is managing. It will *not* restart any of its children if they die,
i.e. we want the container to die and be restarted or noticed by the supervisor.

If one of the programs fails to start dinit will exit with an error. If programs
daemonize dinit will lose track of them.

Dinit has the concept of a *primary* process which is the *last* process listed.
If that process dies dinit will kill the remaining processes and exit. This
allows for cleanups and container restarts. You can also mark all processes as
primary, meaning that if one of them dies, all of them are stopped.

If dinit runs as pid 1 it will cleanup zombies that might be created by any of
the processes.

### Why?

See <https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/>.
But a simpler solution. Get a standard container image and instead of:

    ENTRYPOINT ["/bin/sleep", "80"]

Do:

    ADD dinit /dinit
    ENTRYPOINT ["/dinit", "-r", "/bin/sleep", "80"]

or

    ENTRYPOINT ["/dinit", "-r", "/bin/sleep, "$TIMEOUT"]

Where `$TIMEOUT` will be expanded by `dinit` itself. If you need `-r` as a flag
to a command just escape it with `\-r`, which will be removed by `dinit`.

The last command in the list given to `dinit` will *also* get the arguments given
to `docker run`, so the above sleep can be rewritten like:

    ENTRYPOINT ["/dinit", "-r", "/bin/sleep"]

And then call `docker run .... 80`

Note that the `-start` and `-stop` still take one argument which is split on
whitespace and then executed.


## Socket Interface

When running `dinit` it opens a Unix socket named `/tmp/dinit.sock`. This
enables a text interface that allows for starting extra processes as children of
dinit. The interface is extremely simple: you give it a commandline as you would
normally give to dinit, terminated with a newline.

The string being sent is the command and its arguments: `-r CMD ARG1 ARG2 ... \n`.

The maximum length of the command line that can be sent is 512 characters
including the newline.

With `dinit -submit` you can easily access this functionality:

    % dinit -submit -r sleep 20


## Options

* `maxproc` or `core-fraction`: set GOMAXPROCS to the number of CPUs on the host
  multiplied by `maxproc`, typical values are 0.5 or 1.0. When 0.0 `dinit` will
  not set GOMAXPROCS by itself. If GOMAXPROCS is *already* set in the environment
  this does nothing.
* `start`: run a command when starting up. On any failure, `dinit` exits. The
  complete command must be given as one string, enclosed with quotes.
* `prestop`: run command before passing SIGTERM on to child process(es).
* `prestoptimer`: time in seconds before SIGTERM is sent after the prestop
  command has been started.
* `stop`: run command on exit. The complete command must be given as one string,
  enclosed with quotes.
* `timeout`: time in seconds before SIGKILL is sent after the SIGTERM has been
  sent.
* `primary`: consider all commands primary; if one of them dies then take down
  the other processes.
* `submit`: submit a command line to dinit's socket interface.


## Examples

Start "sleep 2" with dinit, but before you do run `sleep 1`:

    % ./dinit -start "/bin/sleep 1" -r /bin/sleep 2
    2015/07/29 21:49:04 dinit: pid 16759 started: [/bin/sleep 2]
    2015/07/29 21:49:06 dinit: pid 16759, finished: [/bin/sleep 2]
    2015/07/29 21:49:06 dinit: all processes exited, goodbye!

With `-submit` you can start extra processes that will be children of the
original dinit process.

    % dinit -submit -r /bin/sleep 2

Or when dinit is running in a docker container:

    % docker exec a7a55cd8fcf3 /dinit -submit -r /bin/sleep 10


## Environment

The following environment variables are used by dinit:

* DINIT_TIMEOUT: default value to use for timeout.
* DINIT_START: command to run during startup.
* DINIT_PRESTOP: command to run before teardown.
* DINIT_PRESTOPTIMER: seconds to wait before passing on TERM signal.
* DINIT_STOP: command to run during teardown.
* GOMAXPROCS: the GOMAXPROCS for Go programs.

Dinit opens a Unix socket named `/tmp/dinit.sock`.


## See Also

Dinit is partly inspired by
[my_init](https://github.com/phusion/baseimage-docker/blob/master/image/bin/my_init). And init(8).
