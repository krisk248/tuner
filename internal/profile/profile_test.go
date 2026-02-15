package profile

import "testing"

func TestServerValues(t *testing.T) {
	v := ServerValues()
	if v.Governor != "performance" {
		t.Errorf("server governor = %q, want performance", v.Governor)
	}
	if v.Swappiness != 10 {
		t.Errorf("server swappiness = %d, want 10", v.Swappiness)
	}
	if v.TCPCongestion != "bbr" {
		t.Errorf("server tcp congestion = %q, want bbr", v.TCPCongestion)
	}
	if v.DirtyBgRatio != 1 {
		t.Errorf("server dirty_bg_ratio = %d, want 1", v.DirtyBgRatio)
	}
	if v.SchedNVMe != "none" {
		t.Errorf("server nvme sched = %q, want none", v.SchedNVMe)
	}
}

func TestDesktopValues(t *testing.T) {
	v := DesktopValues()
	if v.Governor != "performance" {
		t.Errorf("desktop governor = %q, want performance", v.Governor)
	}
	if v.THPEnabled != "madvise" {
		t.Errorf("desktop THP = %q, want madvise", v.THPEnabled)
	}
	if v.SchedSSD != "kyber" {
		t.Errorf("desktop ssd sched = %q, want kyber", v.SchedSSD)
	}
}

func TestLaptopACValues(t *testing.T) {
	v := LaptopValues(OnAC)
	if v.Governor != "schedutil" {
		t.Errorf("laptop AC governor = %q, want schedutil", v.Governor)
	}
	if !v.TurboOn {
		t.Error("laptop AC turbo should be on")
	}
}

func TestLaptopBatteryValues(t *testing.T) {
	v := LaptopValues(OnBattery)
	if v.Governor != "powersave" {
		t.Errorf("laptop battery governor = %q, want powersave", v.Governor)
	}
	if v.TurboOn {
		t.Error("laptop battery turbo should be off")
	}
	if v.Swappiness != 30 {
		t.Errorf("laptop battery swappiness = %d, want 30", v.Swappiness)
	}
}

func TestIOScheduler(t *testing.T) {
	v := ServerValues()
	tests := []struct {
		diskType string
		want     string
	}{
		{"nvme", "none"},
		{"ssd", "kyber"},
		{"hdd", "bfq"},
		{"unknown", "bfq"},
	}
	for _, tt := range tests {
		got := v.IOScheduler(tt.diskType)
		if got != tt.want {
			t.Errorf("IOScheduler(%q) = %q, want %q", tt.diskType, got, tt.want)
		}
	}
}

func TestForType(t *testing.T) {
	p := ForType(Server)
	if p.Type != Server {
		t.Errorf("ForType(Server).Type = %q, want server", p.Type)
	}
	if p.Values.Governor != "performance" {
		t.Errorf("ForType(Server) governor = %q", p.Values.Governor)
	}

	p = ForType(Desktop)
	if p.Type != Desktop {
		t.Errorf("ForType(Desktop).Type = %q, want desktop", p.Type)
	}
}
