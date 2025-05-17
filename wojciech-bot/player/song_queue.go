package player

import "sync"

// SongQueue represents a thread-safe queue specifically for managing a collection of Song objects.
type SongQueue struct {
	songs []*Song
	mu    sync.Mutex
}

// NewSongQueue creates and returns a new instance of SongQueue with an empty list of songs.
func NewSongQueue() *SongQueue {
	return &SongQueue{
		songs: make([]*Song, 0),
	}
}

// Enqueue adds a new playbackState to the end of the SongQueue in a thread-safe manner.
func (q *SongQueue) Enqueue(song *Song) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.songs = append(q.songs, song)
}

// Dequeue removes and returns the first playbackState in the queue. Returns nil if the queue is empty. Thread-safe.
func (q *SongQueue) Dequeue() *Song {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.songs) == 0 {
		return nil
	}
	song := q.songs[0]
	q.songs = q.songs[1:]
	return song
}

// List returns a copy of the current list of songs in the queue, ensuring thread-safe access.
func (q *SongQueue) List() []*Song {
	q.mu.Lock()
	defer q.mu.Unlock()
	cpy := make([]*Song, len(q.songs))
	copy(cpy, q.songs)
	return cpy
}

// Clear removes all songs from the SongQueue, effectively resetting it to an empty state.
func (q *SongQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.songs = []*Song{}
}

// Length returns the current number of songs in the queue. It is thread-safe and uses a mutex for synchronization.
func (q *SongQueue) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.songs)
}
