const apiBaseUrl = "http://localhost:8080";
let jwtToken = "";
let currentChatroomId = null;
let ws = null;

// Register User
document.getElementById("register-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const username = document.getElementById("register-username").value;
    const password = document.getElementById("register-password").value;

    const res = await fetch(`${apiBaseUrl}/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
    });

    if (res.ok) {
        alert("Registration successful! You can now log in.");
    } else {
        alert("Registration failed.");
    }
});

// Login User
document.getElementById("login-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const username = document.getElementById("login-username").value;
    const password = document.getElementById("login-password").value;

    const res = await fetch(`${apiBaseUrl}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
    });

    if (res.ok) {
        const data = await res.json();
        jwtToken = data.token;
        alert("Login successful!");
        fetchChatrooms();
    } else {
        alert("Login failed.");
    }
});

// Create Chatroom
document.getElementById("create-chatroom-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const name = document.getElementById("chatroom-name").value;

    const res = await fetch(`${apiBaseUrl}/chatroom/create`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${jwtToken}`,
        },
        body: JSON.stringify({ name }),
    });

    if (res.ok) {
        alert("Chatroom created successfully!");
        fetchChatrooms();
    } else {
        alert("Failed to create chatroom.");
    }
});

async function fetchChatrooms() {
    const res = await fetch(`${apiBaseUrl}/chatroom/list`, {
        method: "GET",
        headers: {
            "Authorization": `Bearer ${jwtToken}`,
        },
    });

    if (res.ok) {
        const data = await res.json();
        const chatroomList = document.getElementById("chatroom-list");
        chatroomList.innerHTML = "";

        data.chatrooms.forEach((chatroom) => {
            const li = document.createElement("li");
            const btn = document.createElement("button");
            btn.innerText = chatroom.name;
            btn.onclick = () => joinChatroom(chatroom.id, chatroom.name);
            li.appendChild(btn);
            chatroomList.appendChild(li);
        });
    } else {
        alert("Failed to fetch chatrooms.");
    }
}

async function fetchLastMessages(chatroomId) {
    const res = await fetch(`${apiBaseUrl}/chatroom/messages?chatroom_id=${chatroomId}`, {
        method: "GET",
        headers: {
            "Authorization": `Bearer ${jwtToken}`,
        },
    });

    if (res.ok) {
        const data = await res.json();
        const chatWindow = document.getElementById("chat-window");
        chatWindow.innerHTML = "";

        if (data.messages) {
            data.messages.reverse().forEach((msg) => {
                const messageDiv = document.createElement("div");
                // TODO: Convert 'user_id' to 'username' [it must be changed in the backend as well]
                messageDiv.innerText = `(${new Date(msg.timestamp).toLocaleTimeString()}) User ${msg.user_id}: ${msg.content}`;
                chatWindow.appendChild(messageDiv);
            });
        }
    } else {
        alert("Failed to fetch messages.");
    }
}

// Send Chat Message
document.getElementById("chat-form").addEventListener("submit", (e) => {
    e.preventDefault();
    const messageContent = document.getElementById("chat-message").value;

    if (!currentChatroomId) {
        alert("Please select a chatroom first.");
        return;
    }

    if (!ws || ws.readyState !== WebSocket.OPEN) {
        alert("WebSocket connection is not open. Please refresh the page or select the chatroom again.");
        return;
    }

    const message = { content: messageContent };
    ws.send(JSON.stringify(message));

    document.getElementById("chat-message").value = "";
});

async function joinChatroom(chatroomId, chatroomName) {
    currentChatroomId = chatroomId;
    document.getElementById("chat-window").innerHTML = `<h3>${chatroomName}</h3>`; // Set chatroom name

    await fetchLastMessages(chatroomId);

    setupWebSocket(chatroomId);
}


function setupWebSocket(chatroomId) {
    // Properly close the existing WebSocket before creating a new one
    if (ws) {
        if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
            ws.close();
        }
    }

    // IMPORTANT: This code is for demonstration purposes only.
    // In a real-world application, we should avoid using query parameters for sensitive data such as token.
    // We can use different approaches like using HTTP Handshake or sending the first message with the Auth details.
    ws = new WebSocket(`ws://localhost:8080/ws?chatroom_id=${chatroomId}&token=${jwtToken}`);

    ws.onmessage = (event) => {
        const chatWindow = document.getElementById("chat-window");
        const data = JSON.parse(event.data);
        const messageDiv = document.createElement("div");
        messageDiv.innerText = `(${new Date().toLocaleTimeString()}) ${data.message}`;
        chatWindow.appendChild(messageDiv);
        chatWindow.scrollTop = chatWindow.scrollHeight; // Auto-scroll to bottom
    };

    ws.onerror = (event) => {
        console.error("WebSocket error:", event);
        alert("WebSocket connection failed!");
    };

    ws.onclose = () => {
        console.log("WebSocket connection closed.");
    };
}
