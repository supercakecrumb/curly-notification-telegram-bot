package bot

const (
	startText = `<b>Hi, fella!</b> 👋
I am <b>curly-notifier</b> bot. My purpose is to provide real-time updates on the execution of long-running commands.

To get started, type <code>/help</code> to learn more about how to use me and why I'm useful.`

	helpText = `<b>How to use me:</b>

To get the Bash script, type <code>/getbashscript</code>. It will give you a function that sends a Telegram notification when it finishes execution.

<b>Example usage:</b>

<code>rwtn "echo hello-world"</code>  
This will execute <code>echo hello-world</code> and send you a message:  
<i>Successfully executed: echo hello-world</i>

<code>rwtn "fajlgjdlsjdkf"</code>  
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
        \"telegram_id\": \"{YOUR_TELEGRAM_ID}\",
        \"telegram_password\": \"{YOUR_TELEGRAM_PASSWORD}\"
    }" https://{YOUR_SERVER}/send_notification
}
</pre>`
)
