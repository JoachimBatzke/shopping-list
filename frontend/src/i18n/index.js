import { createI18n } from 'vue-i18n'

const messages = {
  en: {
    // Landing page
    title: 'JORLIST',
    subtitle: 'The simplest shopping list',
    tagline: 'No account. No app store. Works everywhere.',
    feature1_title: 'No Account',
    feature1_desc: 'Start immediately, no signup required',
    feature2_title: 'Works Everywhere',
    feature2_desc: 'Any device with a browser',
    feature3_title: 'Easy Sharing',
    feature3_desc: 'Just send the link to friends',
    step1_title: 'List Name',
    step1_desc: 'Give your list a name',
    step1_placeholder: 'max. 15 letters',
    step2_title: 'App Icon',
    step2_desc: 'Enter or randomize an emoji',
    step2_placeholder: 'Enter emoji',
    random_btn: 'Random',
    step3_title: 'Color',
    step3_desc: 'Choose a color',
    create_btn: 'Create List',
    creating: 'Creating...',
    next: 'Next',
    back: 'Back',

    // List page
    share_btn: 'Share',
    new_list: 'New List',
    add_item: 'Add an item...',
    add_btn: 'Add',
    empty_list: 'Your list is empty. Add some items!',
    loading: 'Loading...',
    delete_btn: 'Delete',

    // Errors
    not_found_title: 'List Not Found',
    not_found_desc: "This list doesn't exist or has been deleted.",
    error_name_required: 'Please enter a list name',
    error_create_failed: 'Failed to create list',
    link_copied: 'Link copied to clipboard!',

    // Footer
    footer: 'JORLIST - No account needed, just share the link'
  },
  de: {
    // Landing page
    title: 'JORLIST',
    subtitle: 'Die einfachste Einkaufsliste',
    tagline: 'Kein Konto. Kein App Store. Funktioniert uberall.',
    feature1_title: 'Kein Konto',
    feature1_desc: 'Sofort starten, keine Anmeldung',
    feature2_title: 'Uberall nutzbar',
    feature2_desc: 'Jedes Gerat mit Browser',
    feature3_title: 'Einfach teilen',
    feature3_desc: 'Link an Freunde senden',
    step1_title: 'Listenname',
    step1_desc: 'Gib deiner Liste einen Namen',
    step1_placeholder: 'max. 15 Zeichen',
    step2_title: 'App-Symbol',
    step2_desc: 'Gib ein Emoji ein oder wurfle',
    step2_placeholder: 'Emoji eingeben',
    random_btn: 'Zufallig',
    step3_title: 'Farbe',
    step3_desc: 'Wahle eine Farbe',
    create_btn: 'Liste erstellen',
    creating: 'Erstelle...',
    next: 'Weiter',
    back: 'Zuruck',

    // List page
    share_btn: 'Teilen',
    new_list: 'Neue Liste',
    add_item: 'Eintrag hinzufugen...',
    add_btn: 'Hinzufugen',
    empty_list: 'Deine Liste ist leer. Fuge Eintrage hinzu!',
    loading: 'Laden...',
    delete_btn: 'Loschen',

    // Errors
    not_found_title: 'Liste nicht gefunden',
    not_found_desc: 'Diese Liste existiert nicht oder wurde geloscht.',
    error_name_required: 'Bitte gib einen Listennamen ein',
    error_create_failed: 'Liste konnte nicht erstellt werden',
    link_copied: 'Link in die Zwischenablage kopiert!',

    // Footer
    footer: 'JORLIST - Kein Konto notig, einfach den Link teilen'
  }
}

// Detect browser language
function getDefaultLocale() {
  const stored = localStorage.getItem('jorlist-locale')
  if (stored && ['en', 'de'].includes(stored)) {
    return stored
  }

  const browserLang = navigator.language.split('-')[0]
  return browserLang === 'de' ? 'de' : 'en'
}

const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: 'en',
  messages
})

export default i18n
