// reveal.js -- Scroll reveal animations, scroll progress bar, back-to-top.
// Vanilla JS, no dependencies. Self-hosted per AGENTS.md.
(function () {
  "use strict";

  // 1. Scroll reveal via IntersectionObserver
  var reveals = document.querySelectorAll(".reveal");
  if (reveals.length > 0 && "IntersectionObserver" in window) {
    var obs = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (entry) {
          if (entry.isIntersecting) {
            entry.target.classList.add("is-visible");
            obs.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.12, rootMargin: "0px 0px -40px 0px" }
    );
    reveals.forEach(function (el) {
      obs.observe(el);
    });
  }

  // 2. Scroll progress bar
  var bar = document.getElementById("scroll-progress");
  // 3. Back-to-top button
  var btn = document.getElementById("back-to-top");

  function onScroll() {
    var scrollTop = window.scrollY || document.documentElement.scrollTop;
    var docHeight = document.documentElement.scrollHeight - window.innerHeight;
    var pct = docHeight > 0 ? (scrollTop / docHeight) * 100 : 0;
    if (bar) bar.style.width = pct + "%";
    if (btn) {
      if (scrollTop > 600) {
        btn.classList.add("is-visible");
      } else {
        btn.classList.remove("is-visible");
      }
    }
    // Glass nav scrolled state
    var nav = document.querySelector(".glass-nav");
    if (nav) nav.classList.toggle("scrolled", scrollTop > 20);
  }

  window.addEventListener("scroll", onScroll, { passive: true });
  onScroll();

  if (btn) {
    btn.addEventListener("click", function () {
      window.scrollTo({ top: 0, behavior: "smooth" });
    });
  }
})();
