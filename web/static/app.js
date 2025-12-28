async function generateKey() {
    return window.crypto.subtle.generateKey(
        { name: "AES-GCM", length: 256 },
        true,
        ["encrypt", "decrypt"]
    );
}

async function encryptData(key, plaintext) {
    const encoder = new TextEncoder();
    const encodedData = encoder.encode(plaintext);

    const iv = window.crypto.getRandomValues(new Uint8Array(12));

    const ciphertext = await window.crypto.subtle.encrypt(
        { name: "AES-GCM", iv: iv },
        key,
        encodedData
    );

    const combined = new Uint8Array(iv.length + ciphertext.byteLength);
    combined.set(iv);
    combined.set(new Uint8Array(ciphertext), iv.length);

    return combined;
}

async function decryptData(rawKeyBytes, encryptedDataBytes) {
    const key = await window.crypto.subtle.importKey(
        "raw",
        rawKeyBytes,
        "AES-GCM",
        true,
        ["decrypt"]
    );

    const iv = encryptedDataBytes.slice(0, 12);
    const data = encryptedDataBytes.slice(12);

    const decryptedBuffer = await window.crypto.subtle.decrypt(
        { name: "AES-GCM", iv: iv },
        key,
        data
    );

    const decoder = new TextDecoder();
    return decoder.decode(decryptedBuffer);
}

function bytesToBase64(bytes) {
    let binary = '';
    const len = bytes.byteLength;
    for (let i = 0; i < len; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary)
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/, '');
}

function base64ToBytes(str) {
    if (str.length % 4 !== 0) {
        str += ('===').slice(0, 4 - (str.length % 4));
    }
    str = str.replace(/-/g, '+').replace(/_/g, '/');
    
    const binary = window.atob(str);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
}

async function createSecret() {
    const errorMsg = document.getElementById('error-msg');
    const secretText = document.getElementById('secret-text').value;
    errorMsg.innerText = "";

    if (!secretText) {
        errorMsg.innerText = "Please enter a secret message.";
        return;
    }

    const hours = parseInt(document.getElementById('ttl-hours').value) || 0;
    const minutes = parseInt(document.getElementById('ttl-minutes').value) || 0;
    const seconds = parseInt(document.getElementById('ttl-seconds').value) || 0;
    const totalSeconds = (hours * 3600) + (minutes * 60) + seconds;

    if (totalSeconds <= 0) {
        errorMsg.innerText = "Duration must be greater than 0.";
        return;
    }

    try {
        const btn = document.getElementById('create-btn');
        btn.disabled = true;
        btn.innerText = "Encrypting...";

        const key = await generateKey();

        const encryptedBytes = await encryptData(key, secretText);

        const payload = {
            data: bytesToBase64(encryptedBytes),
            duration: totalSeconds
        };

        const response = await fetch('/api/v1/secret', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const errData = await response.json();
            throw new Error(errData.error || "Server error");
        }

        const result = await response.json();

        const rawKeyBuffer = await window.crypto.subtle.exportKey("raw", key);
        const keyString = bytesToBase64(new Uint8Array(rawKeyBuffer));

        const fullLink = `${window.location.origin}/view/${result.id}#${keyString}`;

        document.getElementById('create-form').classList.add('hidden');
        document.getElementById('result-area').classList.remove('hidden');
        document.getElementById('result-link').value = fullLink;

    } catch (err) {
        console.error(err);
        errorMsg.innerText = "Error: " + err.message;
        document.getElementById('create-btn').disabled = false;
        document.getElementById('create-btn').innerText = "Create Secret";
    }
}

async function revealSecret() {
    const btn = document.getElementById('reveal-btn');
    const displayArea = document.getElementById('secret-display');
    const contentArea = document.getElementById('secret-content');
    const warningMsg = document.getElementById('warning-msg');

    try {
        btn.innerText = "Decrypting...";
        btn.disabled = true;

        const pathParts = window.location.pathname.split('/');
        const id = pathParts[pathParts.length - 1];
        const keyString = window.location.hash.substring(1);

        if (!id || !keyString) {
            throw new Error("Invalid Link. Missing ID or Decryption Key.");
        }

        const response = await fetch(`/api/v1/secret/${id}`);

        if (response.status === 404) {
            throw new Error("Secret not found or has already been burned.");
        }
        if (response.status === 410) {
            throw new Error("Secret expired.");
        }
	if (!response.ok) {
	    throw new Error("Server communication error.")
	}

        const result = await response.json();

        const keyBytes = base64ToBytes(keyString);
        const encryptedBytes = base64ToBytes(result.data);

        const plaintext = await decryptData(keyBytes, encryptedBytes);

        contentArea.innerText = plaintext;
        displayArea.classList.remove('hidden');
        btn.classList.add('hidden');
        warningMsg.classList.add('hidden');

    } catch (err) {
        console.error(err);
        displayArea.classList.remove('hidden');
        contentArea.innerText = err.message;
        contentArea.style.color = "#ff4d4d";
        btn.classList.add('hidden');
        warningMsg.classList.add('hidden');
    }
}
