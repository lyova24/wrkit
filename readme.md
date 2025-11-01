<div align="center">
    <h3>wrkit</h3>
    <p>
      small, fast task runner driven by YAML files
    </p>

---

#### how to
##### build

<div align="left">

```shell
go mod tidy && go run . build-all
```

</div>

##### install wrkit

<div align="left">

```shell
# replace BUILD_NAME with binary for your platform in ./builds
sudo install -m 755 ./builds/BUILD_NAME /usr/local/bin/wrkit
```

</div>

##### install on-tab completion

<div align="left">

```shell
# run in ~/
wrkit completion bash > ~/.wrkit-completion.sh && echo 'source ~/.wrkit-completion.sh' >> ~/.bashrc
# if using not bash, replace it with fish, powershell or zsh in first part of command
```

</div>

----

##### initialize in directory

<div align="left">

```shell
# this will create wrkit.yaml file for configuration
wrkit -m init
```

</div>

##### run tasks from anywhere

<div align="left">

```shell
# add .wrkit.master.yaml in ~/
touch .wrkit.master.yaml
# tasks from this file will be accessible from any dir
# to ignore .wrkit.master.yaml use --no-master flag
wrkit some-task-name --no-master
# also notice that local wrkit.yaml is more prioritized than .wrkit.master.yaml on conflicts
```

</div>

----

#### learn more
##### wrkit.yaml example

<div align="left">

```yaml
vars: # variables to use
  SLEEP_ALL_SUCCESS_MSG: "all sleep tasks executed successfully!" # can be replaced by --var

tasks:
  sleep-for-2:
    desc: "sleep for 2 seconds" # description
    cmds: |
      sleep 2
      echo "i slept for 2 seconds!"
    parallel: true # is allowed to run parallel

  sleep-for-3:
    desc: "sleep for 3 seconds"
    cmds: | # commands to run
      sleep 3
      echo "i slept for 3 seconds!"
    parallel: true

  sleep-all:
    desc: "run all sleep tasks"
    cmds: |
      echo sleep on the most top level for 2 seconds
      sleep 2
      echo {{.SLEEP_ALL_SUCCESS_MSG}}
    deps: # dependencies list (runs in chaotic order)
      - sleep-for-2
      - sleep-for-3
 ```

</div>

##### read --help for more usage information

<div align="left">

```shell
lyova24@laptop:~$ wrkit --help
wrkit — a small, fast task runner driven by YAML files.

Behavior:
  * If --mode (or -m) is provided, wrkit expects a subcommand (run, list, show, init, version).
    Examples:
      wrkit --mode run task-name
      wrkit -m init

  * If --mode is NOT provided, wrkit treats the first positional argument as a task name
    and runs that task directly:
      wrkit task-name
    This provides a convenient default "run" behavior without typing "run".

Usage:
  wrkit [flags] [task-name]

Flags:
  -c, --concurrency int   Number of tasks to run concurrently (default 4)
      --dry-run           Print what would be done without executing
  -f, --file string       wrkit YAML configuration file (default "wrkit.yaml")
  -h, --help              help for wrkit
  -m, --mode              Enable subcommand mode. When set, use subcommands (run, list, show, init, version).
                          When omitted, the first positional argument is treated as a task name (wrkit <task-name>).
      --no-master         Ignore global ~/.wrkit.master.yaml and use only local wrkit.yaml
  -V, --var stringArray   Variables to pass to templates (key=value). Can be repeated.
  -v, --verbose           Verbose output
```

</div>
</div>
