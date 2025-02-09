package telegram

const (
	startText = `<b>Hi, fella!</b> 👋
I am <b>curly-notifier</b> bot. My purpose is to provide real-time updates on the execution of long-running commands.

To get started, type /help to learn more about how to use me and why I'm useful.`

	helpText = `<b>How to use me:</b>

To get the Bash script, type /getbashscript. It will give you a function that sends a Telegram notification when it finishes execution.

<b>Example usage:</b>

<pre>rwtn "apt upgrade -y && apt update -y"</pre>  
This will execute <code>apt upgrade -y && apt update -y</code> and send you a message:  
<i>Successfully executed: apt upgrade -y && apt update -y</i>

<pre>rwtn "fajlgjdlsjdkf"</pre>  
This will probably send you:  
<i>Failed to execute: fajlgjdlsjdkf</i>`

	bashTemplate = `<b>Add this function to your shell configuration file:</b>
<pre>
# Run with Telegram notification
function rwtn {
    local cmd="$1"

    start_time=$(date +%s)

    (
        set -e
        eval "$cmd"
    )
    exit_code=$?

    end_time=$(date +%s)
    elapsed_time=$(( end_time - start_time ))

    if [ $exit_code -eq 0 ]; then
        TELEGRAM_NOTIFIER_MESSAGE="&lt;b&gt;✅ Successfully executed&lt;/b&gt; (in ${elapsed_time}s): &lt;pre&gt;$cmd&lt;/pre&gt;"
        echo "✅ Successfully executed (in ${elapsed_time}s): $cmd"
    else
        TELEGRAM_NOTIFIER_MESSAGE="&lt;b&gt;Failed to execute&lt;/b&gt; (in ${elapsed_time}s): &lt;pre&gt;$cmd&lt;/pre&gt;"
        echo "❌ Failed to execute: $cmd (in ${elapsed_time}s)"
    fi

    curl -X POST -H "Content-Type: application/json" -d "{
        \"text\": \"$TELEGRAM_NOTIFIER_MESSAGE\",
        \"telegram_id\": \"{{.TelegramID}}\",
        \"password\": \"{{.Password}}\",
        \"parse_mode\": \"HTML\"
    }" {{.APIDomain}}
}
</pre>
`
)
