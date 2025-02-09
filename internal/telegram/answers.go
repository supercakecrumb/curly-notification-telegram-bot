package telegram

const (
	startText = `<b>Hi, fella!</b> 👋
I am <b>curly-notifier</b> bot. My purpose is to provide real-time updates on the execution of long-running commands.

To get started, type <code>/help</code> to learn more about how to use me and why I'm useful.`

	helpText = `<b>How to use me:</b>

To get the Bash script, type /getbashscript. It will give you a function that sends a Telegram notification when it finishes execution.

<b>Example usage:</b>

<pre>rwtn "echo hello-world"</pre>  
This will execute <code>echo hello-world</code> and send you a message:  
<i>Successfully executed: echo hello-world</i>

<pre>rwtn "fajlgjdlsjdkf"</pre>  
This will probably send you:  
<i>Failed to execute: fajlgjdlsjdkf</i>`

	bashScript = `<b>Add this function to your shell configuration file:</b>

<pre>
# Run with Telegram notification
function rwtn {
    local cmd="$1"

    (
        set -e
        eval "$cmd"
    )

    if [ $? -eq 0 ]; then
        TELEGRAM_NOTIFIER_MESSAGE="Successfully executed: $cmd"
        echo "✅ Successfully executed: $cmd"
    else
        TELEGRAM_NOTIFIER_MESSAGE="Failed to execute: $cmd"
        echo "❌ Failed to execute: $cmd"
    fi

    curl -X POST -H "Content-Type: application/json" -d "{
        \"text\": \"$TELEGRAM_NOTIFIER_MESSAGE\",
        \"telegram_id\": \"%v\",
        \"password\": \"%v\"
    }" https://%v/send_notification
}
</pre>`
)
