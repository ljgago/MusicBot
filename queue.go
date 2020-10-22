package main

// QueueGetSong
func (v *VoiceInstance) QueueGetSong() (song Song) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		return v.queue[0]
	}
	return
}

// QueueAdd
func (v *VoiceInstance) QueueAdd(song Song) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	v.queue = append(v.queue, song)
}

// QueueRemoveFirst
func (v *VoiceInstance) QueueRemoveFisrt() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		v.queue = v.queue[1:]
	}
}

// QueueRemoveIndex
func (v *VoiceInstance) QueueRemoveIndex(k int) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 && k <= len(v.queue) {
		v.queue = append(v.queue[:k], v.queue[k+1:]...)
	}
}

// QueueRemoveUser
func (v *VoiceInstance) QueueRemoveUser(user string) {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	queue := v.queue
	v.queue = []Song{}
	if len(v.queue) != 0 {
		for _, q := range queue {
			if q.User != user {
				v.queue = append(v.queue, q)
			}
		}
	}
}

// QueueRemoveLast
func (v *VoiceInstance) QueueRemoveLast() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	if len(v.queue) != 0 {
		v.queue = append(v.queue[:len(v.queue)-1], v.queue[len(v.queue):]...)
	}
}

// QueueClean
func (v *VoiceInstance) QueueClean() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	// hold the actual song in the queue
	v.queue = v.queue[:1]
}

// QueueRemove
func (v *VoiceInstance) QueueRemove() {
	v.queueMutex.Lock()
	defer v.queueMutex.Unlock()
	v.queue = []Song{}
}
