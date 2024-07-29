const { app, BrowserWindow, ipcMain } = require('electron');
const path = require('path');
const { execFile } = require('child_process');
const os = require('os');

function createWindow() {
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      enableRemoteModule: false,
    }
  });

  win.loadFile('index.html');

  win.webContents.openDevTools();
}

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});

ipcMain.handle('deploy-model', async (event, modelContent) => {
  const eicodaPath = os.platform() === 'win32' ? path.join(__dirname, '../EICODA/eicoda.exe') : path.join(__dirname, '../EICODA/eicoda');
  const eicodaWorkingDir = path.join(__dirname, '../EICODA');

  return new Promise((resolve, reject) => {
    const child = execFile(eicodaPath, ['process', '--content', modelContent], { cwd: eicodaWorkingDir }, (error, stdout, stderr) => {
      if (error) {
        reject(stderr);
      } else {
        resolve(stdout.split('\n'));
      }
    });

    child.stdin.write(modelContent);
    child.stdin.end();
  });
});
