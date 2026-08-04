// Minimal native IndexedDB wrapper for offline downloads (Sprint 8) —
// manual eviction only, no automatic LRU cleanup (per docs/roadmap.md).
const DB_NAME = "sonora-offline";
const STORE_NAME = "songs";

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, 1);
    req.onupgradeneeded = () => {
      req.result.createObjectStore(STORE_NAME);
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

export interface DownloadedSong {
  blob: Blob;
  mimeType: string;
  title: string;
  artistName: string;
  sizeBytes: number;
  downloadedAt: number;
}

export async function saveDownload(songId: string, data: DownloadedSong): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, "readwrite");
    tx.objectStore(STORE_NAME).put(data, songId);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
}

export async function getDownload(songId: string): Promise<DownloadedSong | undefined> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, "readonly");
    const req = tx.objectStore(STORE_NAME).get(songId);
    req.onsuccess = () => resolve(req.result as DownloadedSong | undefined);
    req.onerror = () => reject(req.error);
  });
}

export async function deleteDownload(songId: string): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, "readwrite");
    tx.objectStore(STORE_NAME).delete(songId);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
}

export async function listDownloads(): Promise<Array<{ songId: string } & DownloadedSong>> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, "readonly");
    const store = tx.objectStore(STORE_NAME);
    const results: Array<{ songId: string } & DownloadedSong> = [];
    const req = store.openCursor();
    req.onsuccess = () => {
      const cursor = req.result;
      if (cursor) {
        results.push({ songId: String(cursor.key), ...(cursor.value as DownloadedSong) });
        cursor.continue();
      } else {
        resolve(results);
      }
    };
    req.onerror = () => reject(req.error);
  });
}
