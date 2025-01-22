const apiBaseUrl = "http://localhost:8080"; // Backend API base URL
let jwtToken = ""; // Store the JWT token
let currentChatroomId = null; // Active chatroom ID

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
        fetchChatrooms(); // Fetch chatrooms after login
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

// Fetch Available Chatrooms
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
        chatroomList.innerHTML = ""; // Clear existing chatrooms

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

// Join Chatroom and Fetch Last Messages
async function joinChatroom(chatroomId, chatroomName) {
    currentChatroomId = chatroomId;
    document.getElementById("chat-window").innerHTML = `<h3>${chatroomName}</h3>`; // Set chatroom name
    fetchLastMessages(chatroomId);
}

// Fetch Last Messages for the Chatroom
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
        chatWindow.innerHTML = ""; // Clear existing messages

        data.messages.forEach((msg) => {
            const messageDiv = document.createElement("div");
            console.log(msg);
            messageDiv.innerText = `[${new Date(msg.timestamp).toLocaleTimeString()}] ${msg.content}`;
            chatWindow.appendChild(messageDiv);
        });

        chatWindow.scrollTop = chatWindow.scrollHeight; // Auto-scroll to bottom
    } else {
        alert("Failed to fetch messages.");
    }
}

// Send Chat Message
document.getElementById("chat-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const message = document.getElementById("chat-message").value;

    if (!currentChatroomId) {
        alert("Please select a chatroom first.");
        return;
    }

    // Store the message via API
    const res = await fetch(`${apiBaseUrl}/chatroom/post_message`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${jwtToken}`,
        },
        body: JSON.stringify({
            chatroom_id: currentChatroomId,
            content: message,
        }),
    });

    if (res.ok) {
        // Add the message to the chat window
        const chatWindow = document.getElementById("chat-window");
        const messageDiv = document.createElement("div");
        messageDiv.innerText = `You: ${message}`;
        chatWindow.insertBefore(messageDiv, chatWindow.firstChild);

        document.getElementById("chat-message").value = ""; // Clear input
    } else {
        alert("Failed to send message.");
    }
});
