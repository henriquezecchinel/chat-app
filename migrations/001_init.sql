CREATE TABLE IF NOT EXISTS Users (
                                     id SERIAL PRIMARY KEY,
                                     username VARCHAR(255) NOT NULL UNIQUE,
    hashed_password VARCHAR(255) NOT NULL
    );

CREATE TABLE IF NOT EXISTS Chatrooms (
                                         id SERIAL PRIMARY KEY,
                                         name VARCHAR(255) NOT NULL UNIQUE
    );

CREATE TABLE IF NOT EXISTS Messages (
                                        id SERIAL PRIMARY KEY,
                                        chatroom_id INT NOT NULL,
                                        user_id INT NOT NULL,
                                        content TEXT NOT NULL,
                                        timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                        FOREIGN KEY (chatroom_id) REFERENCES Chatrooms(id),
    FOREIGN KEY (user_id) REFERENCES Users(id)
    );