package src

const wrkitYAMLExample = `# wrkit.yaml — example for wrkit
vars:
  SLEEP_ALL_SUCCESS_MSG: "all sleep tasks executed successfully!"

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
      echo {{.SLEEP_ALL_SUCCESS_MSG}}
    deps:
      - sleep-for-2
      - sleep-for-3
`

const cmdRootLongDescription = `wrkit — a small, fast task runner driven by YAML files.

Behavior:
  * If --mode (or -m) is provided, wrkit expects a subcommand (run, list, show, init, version).
    Examples:
      wrkit --mode run task-name
      wrkit -m init

  * If --mode is NOT provided, wrkit treats the first positional argument as a task name
    and runs that task directly:
      wrkit task-name
    This provides a convenient default "run" behavior without typing "run".`
