import { Outlet } from 'react-router-dom';
import Header from './components/Header';
import TorrentEventSync from './components/TorrentEventSync';
import { Suspense } from 'react';

export default function Layout() {
  return (
    <div className="min-h-screen bg-base-200 flex flex-col">
      <TorrentEventSync />
      <Header />

      <main className="flex-1 container mx-auto p-4 max-w-6xl">
        <Suspense fallback={<div className="flex justify-center p-10"><span className="loading loading-spinner loading-lg"></span></div>}>
          <Outlet />
        </Suspense>
      </main>
    </div>
  );
}
