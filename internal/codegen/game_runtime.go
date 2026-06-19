package codegen

// emitGameRuntime emits sprite/input/collision helpers used by app-style games.
func (g *Generator) emitGameRuntime() {
	if !g.usesRaylib || !g.needsGameRuntime {
		return
	}
	if g.usesSpriteRuntime {
		g.emitSpriteRuntime()
	}
	if g.usesInputRuntime {
		g.emitInputRuntime()
	}
	if g.usesCollisionRuntime {
		g.emitCollisionRuntime()
	}
	if g.usesRandomRuntime {
		g.emitRandomRuntime()
	}
	if g.usesVec2Runtime {
		g.emitVec2Runtime()
	}
}

func (g *Generator) emitSpriteRuntime() {
	g.emit(`
/* ── Sprite runtime ── */
typedef struct {
    Texture2D texture;
    Vec2 position;
    Vec2 scale;
    float rotation;
    float speed;
    bool loaded;
    char path[260];
} Sprite;

static Sprite datadream_sprite(const char* path) {
    Sprite s = {0};
    s.scale = (Vec2){1.0f, 1.0f};
    if (path) {
        strncpy(s.path, path, sizeof(s.path) - 1);
    }
    return s;
}

static void datadream_sprite_ensure(Sprite* s) {
    if (s && !s->loaded && s->path[0]) {
        s->texture = LoadTexture(s->path);
        s->loaded = true;
    }
}

static Color datadream_sprite_placeholder_color(Sprite* s) {
    unsigned int hash = 5381;
    if (s && s->path[0]) {
        const char* p = s->path;
        while (*p) { hash = ((hash << 5) + hash) + (unsigned char)(*p++); }
    }
    return (Color){
        (unsigned char)(80 + (hash & 0x7F)),
        (unsigned char)(80 + ((hash >> 8) & 0x7F)),
        (unsigned char)(80 + ((hash >> 16) & 0x7F)),
        255
    };
}

static Rectangle datadream_sprite_bounds(Sprite* s) {
    if (!s) return (Rectangle){0, 0, 0, 0};
    datadream_sprite_ensure(s);
    if (!s->loaded) return (Rectangle){s->position.x, s->position.y, 32, 32};
    return (Rectangle){
        s->position.x,
        s->position.y,
        (float)s->texture.width * s->scale.x,
        (float)s->texture.height * s->scale.y
    };
}

static void datadream_draw_sprite(Sprite* s) {
    if (!s) return;
    datadream_sprite_ensure(s);
    if (s->loaded && s->texture.id > 0) {
        DrawTextureEx(s->texture, s->position, s->rotation, s->scale.x, WHITE);
        return;
    }
    DrawRectangleRec(datadream_sprite_bounds(s), datadream_sprite_placeholder_color(s));
}

static void datadream_sprite_unload(Sprite* s) {
    if (s && s->loaded) {
        UnloadTexture(s->texture);
        s->loaded = false;
        s->texture = (Texture2D){0};
    }
}

`)
}

func (g *Generator) emitInputRuntime() {
	g.emit(`
static int datadream_key_from_name(const char* name) {
    if (!name) return KEY_NULL;
    if (strcmp(name, "left") == 0) return KEY_LEFT;
    if (strcmp(name, "right") == 0) return KEY_RIGHT;
    if (strcmp(name, "up") == 0) return KEY_UP;
    if (strcmp(name, "down") == 0) return KEY_DOWN;
    if (strcmp(name, "space") == 0) return KEY_SPACE;
    if (strcmp(name, "enter") == 0 || strcmp(name, "return") == 0) return KEY_ENTER;
    if (strcmp(name, "escape") == 0 || strcmp(name, "esc") == 0) return KEY_ESCAPE;
    if (strcmp(name, "tab") == 0) return KEY_TAB;
    if (strcmp(name, "backspace") == 0) return KEY_BACKSPACE;
    if (strcmp(name, "delete") == 0) return KEY_DELETE;
    if (strcmp(name, "shift") == 0 || strcmp(name, "lshift") == 0) return KEY_LEFT_SHIFT;
    if (strcmp(name, "rshift") == 0) return KEY_RIGHT_SHIFT;
    if (strcmp(name, "ctrl") == 0 || strcmp(name, "control") == 0 || strcmp(name, "lctrl") == 0) return KEY_LEFT_CONTROL;
    if (strcmp(name, "rctrl") == 0) return KEY_RIGHT_CONTROL;
    if (strcmp(name, "alt") == 0 || strcmp(name, "lalt") == 0) return KEY_LEFT_ALT;
    if (strcmp(name, "ralt") == 0) return KEY_RIGHT_ALT;
    if (strcmp(name, "w") == 0) return KEY_W;
    if (strcmp(name, "a") == 0) return KEY_A;
    if (strcmp(name, "s") == 0) return KEY_S;
    if (strcmp(name, "d") == 0) return KEY_D;
    if (strcmp(name, "q") == 0) return KEY_Q;
    if (strcmp(name, "e") == 0) return KEY_E;
    if (strcmp(name, "f") == 0) return KEY_F;
    if (strcmp(name, "r") == 0) return KEY_R;
    if (strlen(name) == 1) {
        char c = name[0];
        if (c >= 'a' && c <= 'z') return (int)(KEY_A + (c - 'a'));
        if (c >= 'A' && c <= 'Z') return (int)(KEY_A + (c - 'A'));
        if (c >= '0' && c <= '9') return (int)(KEY_ZERO + (c - '0'));
    }
    if (strlen(name) == 2 && name[0] == 'f') {
        int n = name[1] - '0';
        if (n >= 1 && n <= 9) return (int)(KEY_F1 + (n - 1));
    }
    return KEY_NULL;
}

static Vec2 datadream_input_move2d(void) {
    Vec2 move = {0};
    if (IsKeyDown(KEY_RIGHT) || IsKeyDown(KEY_D)) move.x += 1.0f;
    if (IsKeyDown(KEY_LEFT) || IsKeyDown(KEY_A)) move.x -= 1.0f;
    if (IsKeyDown(KEY_DOWN) || IsKeyDown(KEY_S)) move.y += 1.0f;
    if (IsKeyDown(KEY_UP) || IsKeyDown(KEY_W)) move.y -= 1.0f;
    float len = sqrtf(move.x * move.x + move.y * move.y);
    if (len > 0.0f) { move.x /= len; move.y /= len; }
    return move;
}

static Vec2 datadream_input_axis(const char* left, const char* right, const char* up, const char* down) {
    Vec2 move = {0};
    if (IsKeyDown(datadream_key_from_name(right))) move.x += 1.0f;
    if (IsKeyDown(datadream_key_from_name(left))) move.x -= 1.0f;
    if (IsKeyDown(datadream_key_from_name(down))) move.y += 1.0f;
    if (IsKeyDown(datadream_key_from_name(up))) move.y -= 1.0f;
    float len = sqrtf(move.x * move.x + move.y * move.y);
    if (len > 0.0f) { move.x /= len; move.y /= len; }
    return move;
}

static bool datadream_input_pressed(const char* key) {
    int k = datadream_key_from_name(key);
    return k != KEY_NULL && IsKeyPressed(k);
}

static bool datadream_input_down(const char* key) {
    int k = datadream_key_from_name(key);
    return k != KEY_NULL && IsKeyDown(k);
}

static bool datadream_input_released(const char* key) {
    int k = datadream_key_from_name(key);
    return k != KEY_NULL && IsKeyReleased(k);
}

static int datadream_mouse_button_from_name(const char* name) {
    if (!name) return MOUSE_BUTTON_LEFT;
    if (strcmp(name, "left") == 0) return MOUSE_BUTTON_LEFT;
    if (strcmp(name, "right") == 0) return MOUSE_BUTTON_RIGHT;
    if (strcmp(name, "middle") == 0) return MOUSE_BUTTON_MIDDLE;
    return MOUSE_BUTTON_LEFT;
}

static Vec2 datadream_input_mouse(void) {
    return GetMousePosition();
}

static bool datadream_input_mouse_pressed(const char* button) {
    return IsMouseButtonPressed(datadream_mouse_button_from_name(button));
}

static bool datadream_input_mouse_down(const char* button) {
    return IsMouseButtonDown(datadream_mouse_button_from_name(button));
}

static bool datadream_input_mouse_released(const char* button) {
    return IsMouseButtonReleased(datadream_mouse_button_from_name(button));
}

static Vec3 datadream_input_move3d(void) {
    Vec2 xz = datadream_input_move2d();
    Vec3 move = { xz.x, 0.0f, xz.y };
    if (IsKeyDown(KEY_Q)) move.y -= 1.0f;
    if (IsKeyDown(KEY_E)) move.y += 1.0f;
    float len = sqrtf(move.x*move.x + move.y*move.y + move.z*move.z);
    if (len > 0.0f) { move.x /= len; move.y /= len; move.z /= len; }
    return move;
}

static float datadream_input_scroll(void) {
    return GetMouseWheelMoveV().y;
}

`)
}

func (g *Generator) emitCollisionRuntime() {
	g.emit(`
static bool datadream_collision_overlap(Sprite* a, Sprite* b) {
    if (!a || !b) return false;
    return CheckCollisionRecs(datadream_sprite_bounds(a), datadream_sprite_bounds(b));
}

static bool datadream_collision_contains(Sprite* s, Vec2 point) {
    if (!s) return false;
    return CheckCollisionPointRec(point, datadream_sprite_bounds(s));
}

`)
}

func (g *Generator) emitRandomRuntime() {
	g.emit(`
static Vec2 datadream_random_screen_position(void) {
    int w = GetScreenWidth();
    int h = GetScreenHeight();
    if (w < 100) w = 100;
    if (h < 100) h = 100;
    return (Vec2){
        (float)GetRandomValue(32, w - 32),
        (float)GetRandomValue(32, h - 32)
    };
}

static int datadream_random_int(int minVal, int maxVal) {
    if (maxVal < minVal) { int t = minVal; minVal = maxVal; maxVal = t; }
    return GetRandomValue(minVal, maxVal);
}

static float datadream_random_float(float minVal, float maxVal) {
    if (maxVal < minVal) { float t = minVal; minVal = maxVal; maxVal = t; }
    return minVal + (maxVal - minVal) * ((float)GetRandomValue(0, 10000) / 10000.0f);
}

static Vec2 datadream_random_point(Vec2 bounds) {
    int w = (int)bounds.x;
    int h = (int)bounds.y;
    if (w < 64) w = 64;
    if (h < 64) h = 64;
    return (Vec2){
        (float)GetRandomValue(32, w - 32),
        (float)GetRandomValue(32, h - 32)
    };
}

`)
}

func (g *Generator) emitVec2Runtime() {
	g.emit(`
static Vec2 datadream_vec2_add(Vec2 a, Vec2 b) {
    return (Vec2){ a.x + b.x, a.y + b.y };
}

static Vec2 datadream_vec2_mul(Vec2 v, float s) {
    return (Vec2){ v.x * s, v.y * s };
}

`)
}
