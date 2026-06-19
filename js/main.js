// yenikule.com Main Interactive Logic

document.addEventListener('DOMContentLoaded', () => {
  initMobileMenu();
  initScrollAnimations();
  initHeroParallax();
  initContactForm();
});

/**
 * Mobile Navigation Drawer Toggle
 */
function initMobileMenu() {
  const menuButton = document.querySelector('button[aria-controls="mobile-nav"]');
  const mobileNav = document.getElementById('mobile-nav');
  if (!menuButton || !mobileNav) return;

  menuButton.addEventListener('click', () => {
    const isExpanded = menuButton.getAttribute('aria-expanded') === 'true';
    const nextState = !isExpanded;

    menuButton.setAttribute('aria-expanded', String(nextState));
    menuButton.setAttribute('aria-label', nextState ? 'Menüyü kapat' : 'Menüyü aç');

    if (nextState) {
      mobileNav.style.display = 'block';
      // Force repaint to allow transition
      mobileNav.offsetHeight;
      mobileNav.classList.add('open');
      // Change icon to X
      menuButton.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-x"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>`;
    } else {
      mobileNav.classList.remove('open');
      // Wait for transition before hiding display
      setTimeout(() => {
        if (!mobileNav.classList.contains('open')) {
          mobileNav.style.display = 'none';
        }
      }, 300);
      // Change icon to Burger Menu
      menuButton.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-menu"><line x1="4" x2="20" y1="12" y2="12"></line><line x1="4" x2="20" y1="6" y2="6"></line><line x1="4" x2="20" y1="18" y2="18"></line></svg>`;
    }
  });
}

/**
 * Scroll Reveal Animations using IntersectionObserver
 */
function initScrollAnimations() {
  const elements = document.querySelectorAll('.reveal-on-scroll');
  if (elements.length === 0) return;

  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('revealed');
        // Once revealed, we don't need to observe it anymore
        observer.unobserve(entry.target);
      }
    });
  }, {
    threshold: 0.1,
    rootMargin: '0px 0px -50px 0px'
  });

  elements.forEach(element => {
    observer.observe(element);
  });
}

/**
 * Scroll-driven Hero Parallax Effect (Skeleton Building to Completed Building)
 */
function initHeroParallax() {
  const scrollContainer = document.getElementById('hero-scroll-container');
  const skeletonImage = document.getElementById('hero-image-skeleton');
  const finalImage = document.getElementById('hero-image-final');
  const gradientOverlay = document.getElementById('hero-gradient-overlay');
  const heroContent = document.getElementById('hero-text-content');
  const heroActions = document.getElementById('hero-actions-container');
  const scrollIndicator = document.getElementById('hero-scroll-indicator');

  if (!scrollContainer || !skeletonImage || !finalImage) return;

  // Linear interpolation mapping helpers
  function mapRange(value, inMin, inMax, outMin, outMax) {
    if (value <= inMin) return outMin;
    if (value >= inMax) return outMax;
    return outMin + (outMax - outMin) * ((value - inMin) / (inMax - inMin));
  }

  function mapRanges(value, inputPoints, outputPoints) {
    if (value <= inputPoints[0]) return outputPoints[0];
    if (value >= inputPoints[inputPoints.length - 1]) return outputPoints[outputPoints.length - 1];
    for (let i = 0; i < inputPoints.length - 1; i++) {
      if (value >= inputPoints[i] && value <= inputPoints[i+1]) {
        return mapRange(value, inputPoints[i], inputPoints[i+1], outputPoints[i], outputPoints[i+1]);
      }
    }
    return outputPoints[outputPoints.length - 1];
  }

  function handleScroll() {
    const rect = scrollContainer.getBoundingClientRect();
    const scrollContainerHeight = rect.height;
    const viewHeight = window.innerHeight;
    const scrollMax = scrollContainerHeight - viewHeight;

    // Current scroll progress inside the 340vh container (from 0 to 1)
    const progress = Math.max(0, Math.min(1, -rect.top / scrollMax));

    // Map properties based on scroll progress (values from Next.js framer-motion config)
    const scaleSkeleton = mapRanges(progress, [0, 1], [1, 1.22]);
    const scaleFinal = mapRanges(progress, [0, 1], [1.06, 1.26]);
    const opacitySkeleton = mapRanges(progress, [0, 0.3, 0.5], [1, 0.4, 0]);
    const opacityFinal = mapRanges(progress, [0, 0.28, 0.58, 1], [0, 0.3, 1, 1]);
    const ySkeleton = mapRanges(progress, [0, 1], [0, -6]);
    const yFinal = mapRanges(progress, [0, 1], [2, -4]);
    const opacityOverlay = mapRanges(progress, [0, 0.55, 1], [0.58, 0.3,.14]);
    const yText = mapRanges(progress, [0, 1], [0, -14]);
    const opacityIndicator = mapRanges(progress, [0, 0.07], [1, 0]);
    const opacityActions = mapRanges(progress, [0, 0.7, 0.85], [1, 1, 0]);

    // Apply styles to DOM elements
    skeletonImage.style.opacity = opacitySkeleton;
    skeletonImage.style.transform = `scale(${scaleSkeleton}) translateY(${ySkeleton}%)`;

    finalImage.style.opacity = opacityFinal;
    finalImage.style.transform = `scale(${scaleFinal}) translateY(${yFinal}%)`;

    if (gradientOverlay) {
      gradientOverlay.style.opacity = opacityOverlay;
    }

    if (heroContent) {
      heroContent.style.transform = `translateY(${yText}%)`;
    }

    if (heroActions) {
      heroActions.style.opacity = opacityActions;
    }

    if (scrollIndicator) {
      scrollIndicator.style.opacity = opacityIndicator;
    }
  }

  // Bind scroll and resize events
  window.addEventListener('scroll', () => {
    requestAnimationFrame(handleScroll);
  });
  window.addEventListener('resize', handleScroll);

  // Initialize once
  handleScroll();
}

/**
 * Contact Form Submission State Toggle
 */
function initContactForm() {
  const form = document.querySelector('form');
  const formContainer = form ? form.parentElement : null;
  if (!form || !formContainer) return;

  form.addEventListener('submit', (e) => {
    e.preventDefault();

    // Hide the form and show success message
    const successHTML = `
      <div class="flex flex-col items-center justify-center py-16 text-center space-y-4">
        <div class="text-5xl">✅</div>
        <h3 class="heading-md text-dark-bg">Mesajınız İletildi!</h3>
        <p class="text-secondary-text max-w-sm">En kısa sürede size dönüş yapacağız. Bize ulaştığınız için teşekkür ederiz.</p>
        <button type="button" class="text-sm text-beige-accent underline hover:text-dark-bg transition-colors" id="btn-new-message">Yeni mesaj gönder</button>
      </div>
    `;

    // Save initial form HTML for resetting
    const initialFormHTML = formOuterHTML(form);

    formContainer.innerHTML = successHTML;

    // Bind event for sending new message
    const resetBtn = document.getElementById('btn-new-message');
    if (resetBtn) {
      resetBtn.addEventListener('click', () => {
        formContainer.innerHTML = `
          <h2 class="heading-lg mb-8">Bize Yazın</h2>
          ${initialFormHTML}
        `;
        // Re-initialize contact form logic for the new form element
        initContactForm();
      });
    }
  });

  function formOuterHTML(formElement) {
    // Clone form and reset inputs to blank values
    const clone = formElement.cloneNode(true);
    const inputs = clone.querySelectorAll('input, textarea');
    inputs.forEach(input => {
      input.value = '';
    });
    return clone.outerHTML;
  }
}
