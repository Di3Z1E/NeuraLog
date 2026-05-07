import React from 'react';
import { TopBar } from './TopBar';
import { Sidebar } from './Sidebar';
import { LogViewer } from '../logs/LogViewer';

export function Layout() {
  return (
    <div className="app">
      <TopBar />
      <div className="app__body">
        <Sidebar />
        <main className="app__main">
          <LogViewer />
        </main>
      </div>
    </div>
  );
}
