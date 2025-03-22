# Cursor Control-L (CTRL-L) Keyboard Shortcuts in Terminal

When I'm using cursor in Linux, I'm using CTRL-L to clear the terminal screen, but it conflicts with
cursor's default behavior, which is used for "Add to Chat".

The solution is to add a self-defined keyboard shortcut to clear the terminal screen.

Open menu of cursor: File - Settings - Keyboard Shortcuts, and click the "Open Keyboard Shortcuts(JSON)" button,
paste the following code:

```
// Place your key bindings in this file to override the defaults
[
    {
        "key": "ctrl+l",
        "command": "aichat.newchataction",
        "when": "!terminalFocus"
    },
    {
        "key": "ctrl+l",
        "command": "-aichat.newchataction"
    },
    {
        "key": "ctrl+l",
        "command": "workbench.action.terminal.selectCurrentLine",
        "when": "terminalFocus"
    }
]
```

After that, you can use "CTRL-L" to clear the terminal screen.

---

Reference: https://forum.cursor.com/t/change-ctrl-l-to-ctrl-i-control-l-is-for-clear-terminal/15310
