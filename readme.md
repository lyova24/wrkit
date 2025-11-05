<div align="center">

<h3>wrkit</h3>
<p>small, fast task runner driven by YAML files</p>

</div>

---

## 🚀 Quick Start

### 1. Build from source

Make sure you have [Go](https://go.dev/dl/) installed (Go 1.20+ recommended).

```bash
git clone https://github.com/lyova24/wrkit.git
cd wrkit
go mod tidy && go run . build-all
```

After this step, compiled binaries will appear in `./builds/`.

---

### 2. Install binary

After building with `go run . build-all`, you’ll have platform-specific binaries in `./builds/`.  
Choose the one for your system:

#### 🐧 Linux

```bash
sudo install -m 755 ./builds/wrkit.linux.amd64 /usr/local/bin/wrkit
```

#### 🍎 macOS

For Apple Silicon (M1/M2/M3):

```bash
sudo install -m 755 ./builds/wrkit.macos.arm64 /usr/local/bin/wrkit
```

For Intel-based Macs:

```bash
sudo install -m 755 ./builds/wrkit.macos.amd64 /usr/local/bin/wrkit
```

> You may need to grant permission if macOS reports that the binary is from an unidentified developer:
>
> ```bash
> chmod +x /usr/local/bin/wrkit
> xattr -d com.apple.quarantine /usr/local/bin/wrkit
> ```

#### 🪟 Windows

1. Copy the binary to a directory in your PATH, for example:

   ```
   copy .\builds\wrkit.windows.amd64.exe "C:\Program Files\wrkit\wrkit.exe"
   ```
2. Optionally, add `C:\Program Files\wrkit` to your system PATH:

    * Press `Win + R`, type `sysdm.cpl`, go to **Advanced → Environment Variables**
    * Edit **Path**, add the directory, and confirm
3. Now you can run:

   ```powershell
   wrkit
   ```

---

You can now run `wrkit` from anywhere on your system.


### 3. Enable shell autocompletion (optional)

```bash
# Run in ~/
wrkit completion bash > ~/.wrkit-completion.sh
echo 'source ~/.wrkit-completion.sh' >> ~/.bashrc
```

> 🐚 For other shells, replace `bash` with `zsh`, `fish`, or `powershell`.

---

### 4. Initialize in a project

Create a new configuration file in your project directory:

```bash
wrkit -m init
```

This will generate a basic `wrkit.yaml`.

---

### 5. Run a task

Once you’ve added tasks to `wrkit.yaml`, simply run:

```bash
wrkit task-name
```

or, equivalently:

```bash
wrkit -m run task-name
```

---

## 📘 Detailed Guide

### Global tasks

You can define global tasks available from any directory by creating a master config:

```bash
touch ~/.wrkit.master.yaml
```

Tasks in this file are always loaded unless you use the `--no-master` flag:

```bash
wrkit some-task-name --no-master
```

> Local `wrkit.yaml` always has priority over `.wrkit.master.yaml` in case of conflicts.

---

### Variables

Variables can be defined under the `vars:` section or passed via `--var key=value`:

```yaml
vars:
  SLEEP_ALL_SUCCESS_MSG: "all sleep tasks executed successfully!"
```

Command example:

```bash
wrkit sleep-all --var SLEEP_ALL_SUCCESS_MSG="done!"
```

---

### Parallel tasks and dependencies

Each task can have dependencies (`deps:`) and run commands in parallel if `parallel: true` is set.

```yaml
tasks:
  sleep-for-2:
    desc: "sleep for 2 seconds"
    cmds: |
      sleep 2
      echo "i slept for 2 seconds!"
    parallel: true

  sleep-for-3:
    desc: "sleep for 3 seconds"
    cmds: |
      sleep 3
      echo "i slept for 3 seconds!"
    parallel: true

  sleep-all:
    desc: "run all sleep tasks"
    cmds: |
      echo sleeping for 2 seconds at top level
      sleep 2
      echo {{.SLEEP_ALL_SUCCESS_MSG}}
    deps:
      - sleep-for-2
      - sleep-for-3
```

---

### Post-tasks (hooks after main task)

You can specify tasks to run automatically after the main task using the `post:` section.  
Each post-task can have a `when` condition to control when it runs:

- `success` (default): runs only if the main task succeeded
- `fail`: runs only if the main task failed
- `always`: runs regardless of the main task result

Example:

```yaml
tasks:
  build:
    desc: Build the project
    cmds:
      - make build
    post:
      - name: notify
        when: success
      - name: cleanup
        when: always

  notify:
    desc: Notify on build success
    cmds:
      - echo "Build succeeded!"

  cleanup:
    desc: Cleanup after build
    cmds:
      - rm -rf tmp/
```

**How it works:**
- After `build` finishes, `notify` will run only if `build` was successful.
- `cleanup` will always run after `build`, regardless of success or failure.

---

### Log output and task types

During execution, wrkit prints logs with explicit task type labels:

- `[deps-task]` — dependency task (runs before the main task)
- `[main-task]` — the main task you invoked
- `[post-task:success]`, `[post-task:fail]`, `[post-task:always]` — post-tasks, with their trigger condition

Example log fragment:

```
→ [deps-task] sleep-for-2
→ [deps-task] sleep-for-3
→ [main-task] sleep-all
→ [post-task:success] notify
→ [post-task:always] cleanup
```

In verbose mode, command lines are also labeled:

```
[cmd][main-task] echo sleeping for 2 seconds at top level
[cmd][post-task:always] rm -rf tmp/
```

---

## 🔍 CLI Reference

```bash
wrkit --help
```

Example output:

```
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
  -f, --file string       YAML configuration file (default "wrkit.yaml")
  -h, --help              Show help
  -m, --mode              Enable subcommand mode (run, list, show, init, version)
      --no-master         Ignore ~/.wrkit.master.yaml
  -V, --var stringArray   Pass template variables (key=value). Can be repeated.
  -v, --verbose           Verbose output
```

---

## 🧠 Tips

* Local tasks override global ones.
* All commands run in the shell environment by default.
* To preview actions without running them, use `--dry-run`.
* Combine `--concurrency` and `parallel: true` for massive speedups.

---

## 📘 Examples
### Connecting to a Remote VM via Outline

This example demonstrates how to use wrkit in one of the real-world scenarios — connecting to a remote virtual machine through a secure Outline proxy.

```yaml
vars:
  MYVM_USER: some-insane-user
  MYVM_ADDRESS: some-insane-address
  MY_OUTLINE_LINK: ss://key@domain:port/

tasks:
  my-outline:
    cmds: |
      screen -dmS outline sudo ./outline/outline-cli -transport {{.MY_OUTLINE_LINK}}

  my-outline-off:
    cmds: |
      screen -S outline -X quit

  ssh-myvm:
    cmds: |
      ssh {{.MYVM_USER}}@{{.MYVM_ADDRESS}}
    deps:
      - my-outline
    post:
      - name: my-outline-off
        when: always
```

#### What the configuration does

* **`my-outline`** — launches the Outline client in a detached `screen` session to establish a secure VPN/proxy connection using the provided `MY_OUTLINE_LINK`.
* **`my-outline-off`** — stops the running Outline client by terminating the corresponding `screen` session.  
  This task is set as a post-task with `when: always`, so it will always run after `ssh-myvm` finishes (regardless of success or failure).
* **`ssh-myvm`** — connects to the remote VM over SSH using `MYVM_USER` and `MYVM_ADDRESS`.  
  Before execution, this task automatically runs `my-outline` to ensure the SSH connection goes through the secure channel, and after execution, always runs `my-outline-off` to clean up the connection.