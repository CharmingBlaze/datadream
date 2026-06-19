package codegen

func (g *Generator) emitStdlibRuntime() {
	if !g.usesRaylib {
		return
	}
	if g.usesAudioRuntime {
		g.emitAudioRuntime()
	}
	if g.usesInputRuntime && !g.needsGameRuntime {
		g.emitInputRuntime()
	}
}

func (g *Generator) emitAudioRuntime() {
	g.emit(`
typedef struct {
    Sound sound;
    char path[260];
    bool loaded;
} SoundAsset;

static bool _datadream_audio_ready = false;

static void datadream_audio_ensure(void) {
    if (!_datadream_audio_ready) {
        InitAudioDevice();
        _datadream_audio_ready = true;
    }
}

static SoundAsset datadream_sound(const char* path) {
    SoundAsset s = {0};
    if (path) {
        strncpy(s.path, path, sizeof(s.path) - 1);
    }
    return s;
}

static void datadream_sound_ensure(SoundAsset* s) {
    if (s && !s->loaded && s->path[0]) {
        datadream_audio_ensure();
        s->sound = LoadSound(s->path);
        s->loaded = true;
    }
}

static void datadream_audio_play(SoundAsset* s) {
    if (!s) return;
    datadream_sound_ensure(s);
    if (s->loaded) {
        PlaySound(s->sound);
    }
}

static void datadream_audio_stop(SoundAsset* s) {
    if (s && s->loaded) {
        StopSound(s->sound);
    }
}

static void datadream_audio_unload(SoundAsset* s) {
    if (s && s->loaded) {
        UnloadSound(s->sound);
        s->loaded = false;
        s->sound = (Sound){0};
    }
}

static void datadream_audio_shutdown(void) {
    if (_datadream_audio_ready) {
        CloseAudioDevice();
        _datadream_audio_ready = false;
    }
}

`)
}
