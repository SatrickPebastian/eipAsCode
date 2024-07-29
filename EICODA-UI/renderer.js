window.addEventListener('DOMContentLoaded', () => {
    const { ipcRenderer } = window.electron;
  
    let editorInstance = null;
  
    const uploadBtn = document.getElementById('uploadBtn');
    const deployBtn = document.getElementById('deployBtn');
    const fileInput = document.getElementById('fileInput');
    const errorMessage = document.getElementById('errorMessage');
    const modal = document.getElementById('modal');
    const closeModal = document.getElementsByClassName('close')[0];
    const newModelBtn = document.getElementById('newModelBtn');
    const uploadFileBtn = document.getElementById('uploadFileBtn');
  
    uploadBtn.addEventListener('click', () => {
      modal.style.display = 'block';
    });
  
    closeModal.addEventListener('click', () => {
      modal.style.display = 'none';
    });
  
    window.addEventListener('click', (event) => {
      if (event.target == modal) {
        modal.style.display = 'none';
      }
    });
  
    newModelBtn.addEventListener('click', () => {
      modal.style.display = 'none';
      errorMessage.style.display = 'none';
      openEditor('');
    });
  
    uploadFileBtn.addEventListener('click', () => {
      modal.style.display = 'none';
      fileInput.click();
    });
  
    fileInput.addEventListener('change', (event) => {
      const file = event.target.files[0];
      if (file && (file.name.endsWith('.yaml') || file.name.endsWith('.yml'))) {
        const reader = new FileReader();
        reader.onload = (e) => {
          const code = e.target.result;
          errorMessage.style.display = 'none';
          openEditor(code);
        };
        reader.readAsText(file);
      } else {
        errorMessage.textContent = 'Please select a valid .yaml or .yml file.';
        errorMessage.style.display = 'block';
      }
    });
  
    function openEditor(initialCode = '') {
      const editorContainer = document.getElementById('editorContainer');
      editorContainer.style.display = 'block';
  
      if (editorInstance) {
        editorInstance.dispose();
        editorContainer.innerHTML = '';
      }
  
      require.config({ paths: { 'vs': 'https://unpkg.com/monaco-editor/min/vs' }});
      require(['vs/editor/editor.main'], function() {
        editorInstance = monaco.editor.create(editorContainer, {
          value: initialCode,
          language: 'yaml',
          theme: 'vs',
          automaticLayout: true
        });
      });
    }
  });
  