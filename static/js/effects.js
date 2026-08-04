// Premium UI effects - vanilla JS, no dependencies
// Split Text, Count Up, Magnet, Scroll Reveal, Glare Hover, Noise

(function () {
  'use strict';

  // === SPLIT TEXT ===
  // Splits headline into chars/words with staggered entrance
  function initSplitText() {
    var els = document.querySelectorAll('[data-split-text]');
    els.forEach(function (el) {
      var text = el.textContent;
      var words = text.split(' ');
      el.innerHTML = '';
      el.setAttribute('aria-label', text);
      var delay = 0;
      words.forEach(function (word, wi) {
        var wordSpan = document.createElement('span');
        wordSpan.style.display = 'inline-block';
        wordSpan.style.overflow = 'hidden';
        wordSpan.style.verticalAlign = 'top';
        var innerSpan = document.createElement('span');
        innerSpan.className = 'split-char';
        innerSpan.style.display = 'inline-block';
        innerSpan.style.animationDelay = delay + 'ms';
        innerSpan.textContent = word;
        wordSpan.appendChild(innerSpan);
        el.appendChild(wordSpan);
        if (wi < words.length - 1) {
          el.appendChild(document.createTextNode(' '));
        }
        delay += 80;
      });
    });
  }

  // === COUNT UP ===
  // Animates numbers from 0 to target on scroll into view
  function initCountUp() {
    var els = document.querySelectorAll('[data-count-up]');
    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (!entry.isIntersecting) return;
        var el = entry.target;
        var target = parseInt(el.getAttribute('data-count-up'), 10);
        var suffix = el.getAttribute('data-suffix') || '';
        var duration = parseInt(el.getAttribute('data-duration') || '2000', 10);
        var start = null;
        function step(ts) {
          if (!start) start = ts;
          var progress = Math.min((ts - start) / duration, 1);
          var eased = 1 - Math.pow(1 - progress, 3);
          el.textContent = Math.floor(eased * target) + suffix;
          if (progress < 1) requestAnimationFrame(step);
          else el.textContent = target + suffix;
        }
        requestAnimationFrame(step);
        observer.unobserve(el);
      });
    }, { threshold: 0.5 });
    els.forEach(function (el) { observer.observe(el); });
  }

  // === MAGNET ===
  // Elements ease toward cursor then settle back
  function initMagnet() {
    var els = document.querySelectorAll('[data-magnet]');
    els.forEach(function (el) {
      var strength = parseFloat(el.getAttribute('data-magnet-strength') || '0.3');
      el.addEventListener('mousemove', function (e) {
        var rect = el.getBoundingClientRect();
        var x = e.clientX - rect.left - rect.width / 2;
        var y = e.clientY - rect.top - rect.height / 2;
        el.style.transform = 'translate(' + (x * strength) + 'px, ' + (y * strength) + 'px)';
      });
      el.addEventListener('mouseleave', function () {
        el.style.transform = '';
      });
    });
  }

  // === SCROLL REVEAL (enhanced with blur) ===
  // Elements unblur and slide in on scroll
  function initScrollReveal() {
    var els = document.querySelectorAll('[data-scroll-reveal]');
    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.classList.add('scroll-revealed');
          observer.unobserve(entry.target);
        }
      });
    }, { threshold: 0.15, rootMargin: '0px 0px -50px 0px' });
    els.forEach(function (el) { observer.observe(el); });
  }

  // === GLARE HOVER ===
  // Moving glare highlight on hover
  function initGlareHover() {
    var els = document.querySelectorAll('[data-glare]');
    els.forEach(function (el) {
      el.addEventListener('mousemove', function (e) {
        var rect = el.getBoundingClientRect();
        var x = ((e.clientX - rect.left) / rect.width) * 100;
        var y = ((e.clientY - rect.top) / rect.height) * 100;
        el.style.setProperty('--glare-x', x + '%');
        el.style.setProperty('--glare-y', y + '%');
      });
    });
  }

  // === NOISE OVERLAY ===
  // Animated film grain via canvas
  function initNoise() {
    var canvas = document.getElementById('noise-canvas');
    if (!canvas) return;
    var ctx = canvas.getContext('2d');
    var w, h;
    function resize() {
      var rect = canvas.getBoundingClientRect();
      w = rect.width;
      h = rect.height;
      canvas.width = w;
      canvas.height = h;
    }
    resize();
    window.addEventListener('resize', resize);
    function drawNoise() {
      var imgData = ctx.createImageData(w, h);
      var d = imgData.data;
      for (var i = 0; i < d.length; i += 4) {
        var val = Math.random() * 255;
        d[i] = val;
        d[i + 1] = val;
        d[i + 2] = val;
        d[i + 3] = Math.random() * 30;
      }
      ctx.putImageData(imgData, 0, 0);
      requestAnimationFrame(drawNoise);
    }
    drawNoise();
  }

  // === NAVBAR SCROLL ===
  // Transparent -> solid white on scroll
  function initNavScroll() {
    var nav = document.querySelector('.glass-nav');
    if (!nav) return;
    function onScroll() {
      if (window.scrollY > 50) {
        nav.classList.add('scrolled');
      } else {
        nav.classList.remove('scrolled');
      }
    }
    window.addEventListener('scroll', onScroll, { passive: true });
    onScroll();
  }

  // === PARALLAX HERO ===
  // Hero image moves 15% slower than scroll for depth
  function initParallax() {
    var els = document.querySelectorAll('[data-parallax]');
    if (els.length === 0) return;
    function onScroll() {
      var scrollY = window.scrollY;
      els.forEach(function (el) {
        var speed = parseFloat(el.getAttribute('data-parallax') || '0.15');
        el.style.transform = 'translateY(' + (scrollY * speed) + 'px) scale(' + (1 + Math.min(scrollY * 0.0002, 0.05)) + ')';
      });
    }
    window.addEventListener('scroll', onScroll, { passive: true });
    onScroll();
  }

  // === STAGGER REVEAL ===
  // Grids with [data-stagger] cascade children in on viewport entry
  function initStagger() {
    var els = document.querySelectorAll('[data-stagger]');
    if (els.length === 0) return;
    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.classList.add('stagger-visible');
          observer.unobserve(entry.target);
        }
      });
    }, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });
    els.forEach(function (el) { observer.observe(el); });
  }

  // === METRIC LINE DRAW ===
  // .metric-line elements draw left-to-right on viewport entry
  function initMetricLine() {
    var els = document.querySelectorAll('.metric-line');
    if (els.length === 0) return;
    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.classList.add('metric-drawn');
          observer.unobserve(entry.target);
        }
      });
    }, { threshold: 0.5 });
    els.forEach(function (el) { observer.observe(el); });
  }

  // === CUSTOM CURSOR ===
  // Gold ring follows mouse with rAF lag, grows on interactive elements
  function initCustomCursor() {
    if (window.matchMedia('(hover: none), (pointer: coarse)').matches) return;
    var ring = document.createElement('div');
    ring.className = 'cursor-ring cursor-hidden';
    document.body.appendChild(ring);
    var mx = 0, my = 0, rx = 0, ry = 0;
    var shown = false;
    document.addEventListener('mousemove', function (e) {
      mx = e.clientX;
      my = e.clientY;
      if (!shown) { ring.classList.remove('cursor-hidden'); shown = true; }
    });
    document.addEventListener('mouseleave', function () {
      ring.classList.add('cursor-hidden'); shown = false;
    });
    // Grow on interactive elements
    document.addEventListener('mouseover', function (e) {
      if (e.target.closest('a, button, [data-magnet], .btn-chamfer, input, textarea, select')) {
        ring.classList.add('cursor-grow');
      }
    });
    document.addEventListener('mouseout', function (e) {
      if (e.target.closest('a, button, [data-magnet], .btn-chamfer, input, textarea, select')) {
        ring.classList.remove('cursor-grow');
      }
    });
    function tick() {
      rx += (mx - rx) * 0.18;
      ry += (my - ry) * 0.18;
      ring.style.left = rx + 'px';
      ring.style.top = ry + 'px';
      requestAnimationFrame(tick);
    }
    requestAnimationFrame(tick);
  }

  // === SCROLL-DRIVEN VIDEO ===
  // Pinned fullscreen section where scroll controls video.currentTime
  function initScrollVideo() {
    var section = document.querySelector('.scroll-video-section');
    if (!section) return;
    var video = section.querySelector('video');
    var steps = section.querySelectorAll('.scroll-video-step');
    var dots = section.querySelectorAll('.scroll-video-progress span');
    if (!video) return;

    var ticking = false;
    var duration = 0;

    video.addEventListener('loadedmetadata', function () {
      duration = video.duration;
    });
    // If metadata already loaded
    if (video.readyState >= 1) duration = video.duration;
    // Fallback: poll for duration
    if (!duration) {
      var poll = setInterval(function () {
        if (video.duration && isFinite(video.duration)) {
          duration = video.duration;
          clearInterval(poll);
        }
      }, 100);
    }

    function update() {
      var rect = section.getBoundingClientRect();
      var sectionHeight = section.offsetHeight - window.innerHeight;
      var scrolled = Math.max(0, -rect.top);
      var progress = Math.min(1, Math.max(0, scrolled / sectionHeight));

      // Scrub video
      if (duration > 0) {
        video.currentTime = progress * duration;
      }

      // Update steps (divide progress into steps.length segments)
      var stepCount = steps.length;
      if (stepCount > 0) {
        var activeIdx = Math.min(stepCount - 1, Math.floor(progress * stepCount));
        steps.forEach(function (step, i) {
          step.classList.toggle('step-active', i === activeIdx);
        });
        if (dots.length > 0) {
          dots.forEach(function (dot, i) {
            dot.classList.toggle('active', i === activeIdx);
          });
        }
      }

      ticking = false;
    }

    function onScroll() {
      if (!ticking) {
        requestAnimationFrame(update);
        ticking = true;
      }
    }

    window.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('resize', update);
    update();
  }

  // Init all
  function init() {
    initSplitText();
    initCountUp();
    initMagnet();
    initScrollReveal();
    initGlareHover();
    initNoise();
    initNavScroll();
    initParallax();
    initStagger();
    initMetricLine();
    initCustomCursor();
    initScrollVideo();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
