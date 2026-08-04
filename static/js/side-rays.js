// SideRays effect - vanilla WebGL port of reactbits.dev/backgrounds/side-rays
// Self-hosted, no dependencies. Uses raw WebGL with a full-screen triangle.
(function () {
  function hexToRgb(hex) {
    var m = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return m
      ? [parseInt(m[1], 16) / 255, parseInt(m[2], 16) / 255, parseInt(m[3], 16) / 255]
      : [1, 1, 1];
  }

  function initSideRays(container, opts) {
    opts = opts || {};
    var speed = opts.speed || 2.5;
    var rayColor1 = opts.rayColor1 || '#d4af6a';
    var rayColor2 = opts.rayColor2 || '#96c8ff';
    var intensity = opts.intensity || 2;
    var spread = opts.spread || 2;
    var origin = opts.origin || 'top-right';
    var tilt = opts.tilt || 0;
    var saturation = opts.saturation || 1.5;
    var blend = opts.blend || 0.75;
    var falloff = opts.falloff || 2.0;
    var opacity = opts.opacity || 0.6;

    var flipX, flipY;
    switch (origin) {
      case 'top-left': flipX = 1; flipY = 0; break;
      case 'bottom-right': flipX = 0; flipY = 1; break;
      case 'bottom-left': flipX = 1; flipY = 1; break;
      default: flipX = 0; flipY = 0;
    }

    var canvas = document.createElement('canvas');
    canvas.style.width = '100%';
    canvas.style.height = '100%';
    canvas.style.display = 'block';
    container.appendChild(canvas);

    var gl = canvas.getContext('webgl', { alpha: true, premultipliedAlpha: false });
    if (!gl) return;

    var vertSrc =
      'attribute vec2 position;\n' +
      'void main() {\n' +
      '  gl_Position = vec4(position, 0.0, 1.0);\n' +
      '}';

    var fragSrc =
      'precision highp float;\n' +
      'uniform float iTime;\n' +
      'uniform vec2 iResolution;\n' +
      'uniform float iSpeed;\n' +
      'uniform vec3 iRayColor1;\n' +
      'uniform vec3 iRayColor2;\n' +
      'uniform float iIntensity;\n' +
      'uniform float iSpread;\n' +
      'uniform float iFlipX;\n' +
      'uniform float iFlipY;\n' +
      'uniform float iTilt;\n' +
      'uniform float iSaturation;\n' +
      'uniform float iBlend;\n' +
      'uniform float iFalloff;\n' +
      'uniform float iOpacity;\n' +
      'float rayStrength(vec2 raySource, vec2 rayRefDirection, vec2 coord, float seedA, float seedB, float speed) {\n' +
      '  vec2 sourceToCoord = coord - raySource;\n' +
      '  float cosAngle = dot(normalize(sourceToCoord), rayRefDirection);\n' +
      '  return clamp(\n' +
      '    (0.45 + 0.15 * sin(cosAngle * seedA + iTime * speed)) +\n' +
      '    (0.3 + 0.2 * cos(-cosAngle * seedB + iTime * speed)),\n' +
      '    0.0, 1.0) *\n' +
      '    clamp((iResolution.x - length(sourceToCoord)) / iResolution.x, 0.5, 1.0);\n' +
      '}\n' +
      'void main() {\n' +
      '  vec2 fragCoord = gl_FragCoord.xy;\n' +
      '  if (iFlipX > 0.5) fragCoord.x = iResolution.x - fragCoord.x;\n' +
      '  if (iFlipY > 0.5) fragCoord.y = iResolution.y - fragCoord.y;\n' +
      '  vec2 coord = vec2(fragCoord.x, iResolution.y - fragCoord.y);\n' +
      '  vec2 rayPos = vec2(iResolution.x * 1.1, -0.5 * iResolution.y);\n' +
      '  float tiltRad = iTilt * 3.14159265 / 180.0;\n' +
      '  float cs = cos(tiltRad);\n' +
      '  float sn = sin(tiltRad);\n' +
      '  vec2 rel = coord - rayPos;\n' +
      '  vec2 tiltedCoord = vec2(rel.x * cs - rel.y * sn, rel.x * sn + rel.y * cs) + rayPos;\n' +
      '  float halfSpread = iSpread * 0.275;\n' +
      '  vec2 rayRefDir1 = normalize(vec2(cos(0.785398 + halfSpread), sin(0.785398 + halfSpread)));\n' +
      '  vec2 rayRefDir2 = normalize(vec2(cos(0.785398 - halfSpread), sin(0.785398 - halfSpread)));\n' +
      '  vec4 rays1 = vec4(iRayColor1, 1.0) * rayStrength(rayPos, rayRefDir1, tiltedCoord, 36.2214, 21.11349, iSpeed);\n' +
      '  vec4 rays2 = vec4(iRayColor2, 1.0) * rayStrength(rayPos, rayRefDir2, tiltedCoord, 22.3991, 18.0234, iSpeed * 0.2);\n' +
      '  vec4 color = rays1 * (1.0 - iBlend) * 0.9 + rays2 * iBlend * 0.9;\n' +
      '  float distanceToLight = length(fragCoord.xy - vec2(rayPos.x, iResolution.y - rayPos.y)) / iResolution.y;\n' +
      '  float brightness = iIntensity * 0.4 / pow(max(distanceToLight, 0.001), iFalloff);\n' +
      '  color.rgb *= brightness;\n' +
      '  float gray = dot(color.rgb, vec3(0.299, 0.587, 0.114));\n' +
      '  color.rgb = mix(vec3(gray), color.rgb, iSaturation);\n' +
      '  color.a = max(color.r, max(color.g, color.b)) * iOpacity;\n' +
      '  gl_FragColor = color;\n' +
      '}';

    function compile(type, src) {
      var s = gl.createShader(type);
      gl.shaderSource(s, src);
      gl.compileShader(s);
      return s;
    }

    var program = gl.createProgram();
    gl.attachShader(program, compile(gl.VERTEX_SHADER, vertSrc));
    gl.attachShader(program, compile(gl.FRAGMENT_SHADER, fragSrc));
    gl.linkProgram(program);
    gl.useProgram(program);

    // Full-screen triangle
    var buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW);
    var posLoc = gl.getAttribLocation(program, 'position');
    gl.enableVertexAttribArray(posLoc);
    gl.vertexAttribPointer(posLoc, 2, gl.FLOAT, false, 0, 0);

    var uniforms = {
      iTime: gl.getUniformLocation(program, 'iTime'),
      iResolution: gl.getUniformLocation(program, 'iResolution'),
      iSpeed: gl.getUniformLocation(program, 'iSpeed'),
      iRayColor1: gl.getUniformLocation(program, 'iRayColor1'),
      iRayColor2: gl.getUniformLocation(program, 'iRayColor2'),
      iIntensity: gl.getUniformLocation(program, 'iIntensity'),
      iSpread: gl.getUniformLocation(program, 'iSpread'),
      iFlipX: gl.getUniformLocation(program, 'iFlipX'),
      iFlipY: gl.getUniformLocation(program, 'iFlipY'),
      iTilt: gl.getUniformLocation(program, 'iTilt'),
      iSaturation: gl.getUniformLocation(program, 'iSaturation'),
      iBlend: gl.getUniformLocation(program, 'iBlend'),
      iFalloff: gl.getUniformLocation(program, 'iFalloff'),
      iOpacity: gl.getUniformLocation(program, 'iOpacity'),
    };

    var c1 = hexToRgb(rayColor1);
    var c2 = hexToRgb(rayColor2);

    gl.uniform1f(uniforms.iSpeed, speed);
    gl.uniform3f(uniforms.iRayColor1, c1[0], c1[1], c1[2]);
    gl.uniform3f(uniforms.iRayColor2, c2[0], c2[1], c2[2]);
    gl.uniform1f(uniforms.iIntensity, intensity);
    gl.uniform1f(uniforms.iSpread, spread);
    gl.uniform1f(uniforms.iFlipX, flipX);
    gl.uniform1f(uniforms.iFlipY, flipY);
    gl.uniform1f(uniforms.iTilt, tilt);
    gl.uniform1f(uniforms.iSaturation, saturation);
    gl.uniform1f(uniforms.iBlend, blend);
    gl.uniform1f(uniforms.iFalloff, falloff);
    gl.uniform1f(uniforms.iOpacity, opacity);

    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

    var dpr = Math.min(window.devicePixelRatio || 1, 2);

    function resize() {
      var w = container.clientWidth;
      var h = container.clientHeight;
      canvas.width = w * dpr;
      canvas.height = h * dpr;
      gl.viewport(0, 0, canvas.width, canvas.height);
      gl.uniform2f(uniforms.iResolution, canvas.width, canvas.height);
    }

    window.addEventListener('resize', resize);
    resize();

    var animId = null;
    var startTime = performance.now();

    function loop() {
      var t = (performance.now() - startTime) * 0.001;
      gl.uniform1f(uniforms.iTime, t);
      gl.drawArrays(gl.TRIANGLES, 0, 3);
      animId = requestAnimationFrame(loop);
    }

    // Pause when not visible (IntersectionObserver)
    var observer = new IntersectionObserver(function (entries) {
      if (entries[0].isIntersecting) {
        if (!animId) loop();
      } else {
        if (animId) {
          cancelAnimationFrame(animId);
          animId = null;
        }
      }
    }, { threshold: 0.05 });
    observer.observe(container);

    // Respect reduced motion
    if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      gl.uniform1f(uniforms.iSpeed, 0);
    }

    loop();
  }

  window.initSideRays = initSideRays;
})();
