// Beams animation - Canvas 2D, no dependencies
// Inspired by React Bits Beams, adapted for vanilla JS
(function () {
  var canvas, ctx, w, h, beams = [], time = 0, raf;

  var config = {
    beamNumber: 14,
    beamWidth: 3,
    beamHeight: 0.85,
    speed: 0.8,
    lightColor: '234, 200, 8',
    baseColor: '3, 22, 54'
  };

  function init() {
    canvas = document.getElementById('beams-canvas');
    if (!canvas) return;
    ctx = canvas.getContext('2d');
    resize();
    createBeams();
    animate();
    window.addEventListener('resize', function () {
      resize();
      createBeams();
    });
  }

  function resize() {
    var dpr = window.devicePixelRatio || 1;
    var rect = canvas.getBoundingClientRect();
    w = rect.width;
    h = rect.height;
    canvas.width = w * dpr;
    canvas.height = h * dpr;
    ctx.scale(dpr, dpr);
  }

  function noise(x, y, t) {
    return Math.sin(x * 0.5 + t) * Math.cos(y * 0.3 + t * 0.7) * 0.5 +
           Math.sin(x * 1.2 + t * 1.3) * 0.3 +
           Math.cos(y * 0.8 - t * 0.5) * 0.2;
  }

  function createBeams() {
    beams = [];
    var spacing = w / config.beamNumber;
    for (var i = 0; i < config.beamNumber; i++) {
      beams.push({
        x: i * spacing + spacing / 2,
        baseX: i * spacing + spacing / 2,
        width: config.beamWidth + Math.random() * 2,
        offset: Math.random() * 100,
        phase: Math.random() * Math.PI * 2
      });
    }
  }

  function animate() {
    time += 0.016 * config.speed;

    // Dark background
    ctx.fillStyle = 'rgba(' + config.baseColor + ', 1)';
    ctx.fillRect(0, 0, w, h);

    // Subtle radial glow center
    var grad = ctx.createRadialGradient(w / 2, h / 2, 0, w / 2, h / 2, w * 0.6);
    grad.addColorStop(0, 'rgba(' + config.lightColor + ', 0.04)');
    grad.addColorStop(1, 'rgba(' + config.baseColor + ', 0)');
    ctx.fillStyle = grad;
    ctx.fillRect(0, 0, w, h);

    // Draw beams
    for (var i = 0; i < beams.length; i++) {
      var b = beams[i];
      var n = noise(b.offset, 0, time);
      var x = b.baseX + n * 30;
      var beamH = h * config.beamHeight;
      var yStart = (h - beamH) / 2;
      var opacity = 0.15 + Math.abs(n) * 0.25;

      // Beam gradient (vertical fade)
      var beamGrad = ctx.createLinearGradient(0, yStart, 0, yStart + beamH);
      beamGrad.addColorStop(0, 'rgba(' + config.lightColor + ', 0)');
      beamGrad.addColorStop(0.5, 'rgba(' + config.lightColor + ', ' + opacity + ')');
      beamGrad.addColorStop(1, 'rgba(' + config.lightColor + ', 0)');

      ctx.fillStyle = beamGrad;
      ctx.fillRect(x - b.width / 2, yStart, b.width, beamH);

      // Glow halo around beam
      var glowGrad = ctx.createLinearGradient(x - 20, 0, x + 20, 0);
      glowGrad.addColorStop(0, 'rgba(' + config.lightColor + ', 0)');
      glowGrad.addColorStop(0.5, 'rgba(' + config.lightColor + ', ' + (opacity * 0.3) + ')');
      glowGrad.addColorStop(1, 'rgba(' + config.lightColor + ', 0)');
      ctx.fillStyle = glowGrad;
      ctx.fillRect(x - 20, yStart, 40, beamH);
    }

    // Noise overlay (subtle grain)
    if (Math.random() > 0.5) {
      var imgData = ctx.getImageData(0, 0, Math.min(w, 1920), Math.min(h, 800));
      var d = imgData.data;
      for (var j = 0; j < d.length; j += 4) {
        var grain = (Math.random() - 0.5) * 8;
        d[j] += grain;
        d[j + 1] += grain;
        d[j + 2] += grain;
      }
      ctx.putImageData(imgData, 0, 0);
    }

    raf = requestAnimationFrame(animate);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
