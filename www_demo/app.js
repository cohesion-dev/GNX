let API_BASE_URL = localStorage.getItem('api_base_url') || 'http://localhost:8000';

document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('api-base-url').value = API_BASE_URL;
});

function saveConfig() {
    API_BASE_URL = document.getElementById('api-base-url').value;
    localStorage.setItem('api_base_url', API_BASE_URL);
    alert('配置已保存');
}

function displayResult(elementId, data) {
    document.getElementById(elementId).textContent = JSON.stringify(data, null, 2);
}

async function getComics() {
    try {
        const response = await fetch(`${API_BASE_URL}/comics/`);
        const data = await response.json();
        displayResult('comics-result', data);
    } catch (error) {
        displayResult('comics-result', { error: error.message });
    }
}

async function uploadNovel() {
    const title = document.getElementById('comic-title').value;
    const userPrompt = document.getElementById('user-prompt').value;
    const fileInput = document.getElementById('novel-file');
    const file = fileInput.files[0];

    if (!title || !file) {
        alert('请填写标题并选择文件');
        return;
    }

    const formData = new FormData();
    formData.append('title', title);
    formData.append('user_prompt', userPrompt);
    formData.append('file', file);

    try {
        const response = await fetch(`${API_BASE_URL}/comics/`, {
            method: 'POST',
            body: formData
        });
        const data = await response.json();
        displayResult('upload-result', data);
    } catch (error) {
        displayResult('upload-result', { error: error.message });
    }
}

async function getComicDetail() {
    const comicId = document.getElementById('comic-id').value;
    if (!comicId) {
        alert('请输入漫画ID');
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/comics/${comicId}/`);
        const data = await response.json();
        displayResult('detail-result', data);
    } catch (error) {
        displayResult('detail-result', { error: error.message });
    }
}

async function createSection() {
    const comicId = document.getElementById('section-comic-id').value;
    const title = document.getElementById('section-title').value;
    const content = document.getElementById('section-content').value;

    if (!comicId || !title || !content) {
        alert('请填写所有字段');
        return;
    }

    const formData = new FormData();
    formData.append('title', title);
    formData.append('content', content);

    try {
        const response = await fetch(`${API_BASE_URL}/comics/${comicId}/sections/`, {
            method: 'POST',
            body: formData
        });
        const data = await response.json();
        displayResult('section-result', data);
    } catch (error) {
        displayResult('section-result', { error: error.message });
    }
}

async function getSectionDetail() {
    const comicId = document.getElementById('section-detail-comic-id').value;
    const sectionId = document.getElementById('section-detail-id').value;

    if (!comicId || !sectionId) {
        alert('请输入漫画ID和章节ID');
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/comics/${comicId}/sections/${sectionId}/`);
        const data = await response.json();
        displayResult('section-detail-result', data);
    } catch (error) {
        displayResult('section-detail-result', { error: error.message });
    }
}

async function getImageUrl() {
    const imageId = document.getElementById('image-id').value;
    if (!imageId) {
        alert('请输入图片ID');
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/images/${imageId}/url`);
        const data = await response.json();
        displayResult('image-result', data);
        
        if (data.code === 200 && data.data.url) {
            const imgPreview = document.getElementById('image-preview');
            imgPreview.innerHTML = `<img src="${data.data.url}" alt="Preview" style="max-width: 100%; margin-top: 10px;">`;
        }
    } catch (error) {
        displayResult('image-result', { error: error.message });
    }
}

function playTTS() {
    const ttsId = document.getElementById('tts-id').value;
    if (!ttsId) {
        alert('请输入TTS ID');
        return;
    }

    const audioUrl = `${API_BASE_URL}/tts/${ttsId}`;
    const audioPlayer = document.getElementById('audio-player');
    audioPlayer.innerHTML = `
        <audio controls style="margin-top: 10px;">
            <source src="${audioUrl}" type="audio/mpeg">
            您的浏览器不支持音频播放。
        </audio>
    `;
}
