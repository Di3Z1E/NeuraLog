import { useState } from 'react';
import { TopBar } from './TopBar';
import { Sidebar } from './Sidebar';
import { LogViewer } from '../logs/LogViewer';
import { SettingsModal } from '../settings/SettingsModal';

export function Layout() {
  const [settingsOpen, setSettingsOpen] = useState(false);

  return (
    <div className="app">
      <TopBar onSettings={() => setSettingsOpen(true)} />
      <div className="app__body">
        <Sidebar />
        <main className="app__main">
          <LogViewer />
        </main>
      </div>
      {settingsOpen && <SettingsModal onClose={() => setSettingsOpen(false)} />}
    </div>
  );
}
