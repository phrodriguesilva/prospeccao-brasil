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

  // Init all
  function init() {
    initSplitText();
    initCountUp();
    initMagnet();
    initScrollReveal();
    initGlareHover();
    initNoise();
    initNavScroll();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
