(() => {
  function dial() {
    const recId = location.pathname.split("/");
    const conn = new WebSocket(
      `ws://${location.host}/subscribe/${recId[recId.length - 1]}`,
    );

    conn.addEventListener("close", (ev) => {
      appendLog(
        `WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`,
        true,
      );
      if (ev.code !== 1001) {
        appendLog("Reconnecting in 1s", true);
        setTimeout(dial, 1000);
      }
    });
    conn.addEventListener("open", (ev) => {
      console.info("websocket connected");
    });

    // This is where we handle messages received.
    conn.addEventListener("message", (ev) => {
      if (typeof ev.data !== "string") {
        console.error("unexpected message type", typeof ev.data);
        return;
      }
      const m = JSON.parse(ev.data);
      console.log(m);
      const p = appendLog(m);
      p.scrollIntoView();
      p.scrollTop = p.scrollHeight;
    });
  }
  dial();

  const messageLog = document.getElementById("message-log");
  const publishForm = document.getElementById("publish-form");
  const messageInput = document.getElementById("message-input");
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  function createMessage(m){
    const msgWrapper = document.createElement("div")
    msgWrapper.className = "flex items-start mb-4 p-3 rounded-lg shadow-md"

    const profileImage = document.createElement("div")
    profileImage.className = "w-10 h-10 bg-gray-400 rounded-full flex-shrink-0 mr-3"

    msgWrapper.append(profileImage)

    const msgBody = document.createElement("div")
    msgBody.className = "flex-1"

    const usernameContainer = document.createElement("div")
    usernameContainer.className = "flex items-center mb-1"

    const username = document.createElement("span")
    username.className = "font-semibold text-love"
    username.innerText = "Elvis"

    usernameContainer.append(username)

    msgBody.append(usernameContainer)

    const msg= document.createElement("p")
    msg.className = "text-text"
    msg.innerText = m.body
    msgBody.append(msg)

    msgWrapper.append(msgBody)
    return msgWrapper
  }
  function scrollToBottom(){
    messageLog.scrollTop = messageLog.scrollHeight;
  }
  window.onload = scrollToBottom
  // appendLog appends the passed text to messageLog.
  function appendLog(msg, error) {
    const m = createMessage(msg)
    // Adding a timestamp to each message makes the log easier to read.
    // p.innerText = `${new Date().toLocaleTimeString()}: ${text}`;
    // p.innerText = text
    // if (error) {
    //   p.style.color = "red";
    //   p.style.fontStyle = "bold";
    // }
    messageLog.append(m);
    return m;
  }

  // onsubmit publishes the message from the user when the form is submitted.
  publishForm.onsubmit = async (ev) => {
    ev.preventDefault();

    const msg = messageInput.value;
    if (msg === "") {
      return;
    }
    messageInput.value = "";

    expectingMessage = true;
    const formData = new FormData();
    const csrfToken = document.querySelector("[name=csrf_token]").value;
    formData.append("message", msg);
    formData.append("csrf_token", csrfToken);
    formData.append("senderId", window.location.pathname.split("/")[2]);
    formData.append("receiverId", window.location.pathname.split("/")[2]);
    formData.append("timezone", timeZone);
    try {
      const resp = await fetch("/publish", {
        method: "POST",
        body: formData,
      });
      if (resp.status !== 202) {
        throw new Error(
          `Unexpected HTTP Status ${resp.status} ${resp.statusText}`,
        );
      }
    } catch (err) {
      appendLog(`Publish failed: ${err.message}`, true);
    }
  };
})();
