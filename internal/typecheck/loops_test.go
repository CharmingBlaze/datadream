package typecheck

import (
	"strings"
	"testing"
)

func hasWarning(errs []Error, substr string) bool {
	for _, e := range errs {
		if e.Warning && strings.Contains(e.Message, substr) {
			return true
		}
	}
	return false
}

func hasErrorNotWarning(errs []Error, substr string) bool {
	for _, e := range errs {
		if !e.Warning && strings.Contains(e.Message, substr) {
			return true
		}
	}
	return false
}

func TestInfiniteLoopInUpdateRejected(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
update {
    loop {
        let x = 1;
    }
}`)
	if !hasErrorNotWarning(errs, "infinite `loop` inside a per-frame block") {
		t.Fatalf("expected infinite loop error, got %v", errs)
	}
}

func TestLoopWithBreakAllowedInUpdate(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
update {
    loop {
        if true { break; }
    }
}`)
	for _, e := range errs {
		if !e.Warning && strings.Contains(e.Message, "infinite `loop`") {
			t.Fatalf("unexpected infinite loop error: %v", errs)
		}
	}
}

func TestLoopAllowedInStartBlock(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
start {
    loop {
        let x = 1;
    }
}`)
	for _, e := range errs {
		if strings.Contains(e.Message, "infinite `loop`") {
			t.Fatalf("start block loop should be allowed, got %v", errs)
		}
	}
}

func TestForRangeRuntimeBoundWarning(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
let count = 10;
update {
    for i in 0..count { }
}`)
	if !hasWarning(errs, "loop upper bound is not known at compile time") {
		t.Fatalf("expected range bound warning, got %v", errs)
	}
}

func TestNestedEntityLoopWarning(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
entity Enemy { }
entity Bullet { }
update {
    for e in Enemy {
        for b in Bullet { }
    }
}`)
	if !hasWarning(errs, "nested `for x in") {
		t.Fatalf("expected nested entity warning, got %v", errs)
	}
}

func TestAllocationInPerFrameLoopWarning(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
update {
    for i in 0..5 {
        let label = "Enemy_" + "1";
    }
}`)
	if !hasWarning(errs, "allocation inside a per-frame loop") {
		t.Fatalf("expected allocation warning, got %v", errs)
	}
}

func TestArrayPushInLoopWarning(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
let bullets: Array<int>;
update {
    for i in 0..3 {
        bullets.push(1);
    }
}`)
	if !hasWarning(errs, "growing `bullets` with `.push()`") {
		t.Fatalf("expected array push warning, got %v", errs)
	}
}

func TestDrawMutationWarning(t *testing.T) {
	errs := parseCheck(t, `app "T";
window { size: 100, 100; title: "T"; }
entity Enemy { speed: float; }
draw {
    for e in Enemy {
        e.speed = 1.0;
    }
}`)
	if !hasWarning(errs, "mutating entity or game state inside `draw`") {
		t.Fatalf("expected draw mutation warning, got %v", errs)
	}
}
