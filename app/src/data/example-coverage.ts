import type { ProfessionalCoverage } from '../types/coverage-professional'

export const EXAMPLE_PROFESSIONAL_COVERAGE: ProfessionalCoverage = {
  id: 'the-arrangement-professional',
  title: 'THE ARRANGEMENT',
  author: 'Maren Altman',
  genre: ['Psychological Drama', 'Period Thriller'],
  setting: '1960s Italy (Amalfi Coast)',
  page_count: 110,
  format: 'feature',
  analysis_date: '2025-11-20',
  analyst_names: ['Research Agent', 'Comprehensive Analyst', 'Author Profile Agent', 'Scene Analyst 1-4'],

  logline: 'A young American actress becomes the latest muse of a renowned European director at his isolated Italian villa, where the boundaries between artistic mentorship and psychological manipulation blur as she transforms from uncertain ingenue to complicit star.',

  synopsis: {
    act_one: `Detective Marcus Kane is a burned-out LAPD veteran haunted by the death of his partner two years ago. His marriage has fallen apart, and his relationship with his 23-year-old daughter Emma, a CDC scientist, is virtually nonexistent.\n\nWhen a series of mysterious deaths occur across the city, Emma discovers a pattern that suggests a weaponized pathogen. She tries to warn her father, but he dismisses her concerns.`,
    act_two: `The situation escalates when Emma is kidnapped by a sophisticated terrorist cell led by the enigmatic Victor Kaine, a former bioweapons researcher with a vendetta against the government. Marcus is forced to confront his failures as both a detective and a father.\n\nWorking with Emma's research partner and an FBI task force, Marcus uncovers Kaine's plan to release the pathogen during a major public event. The clock is ticking, and Marcus must navigate both the criminal underworld and his own demons to find his daughter.`,
    act_three: `Marcus infiltrates Kaine's compound and rescues Emma, but not before learning that she's been exposed to the pathogen. With only hours before symptoms appear, they must work together to stop Kaine's plan and find the antidote.`,
    resolution: `The climax takes place at the event venue, where Marcus and Emma confront Kaine in a tense standoff that tests both their courage and their rekindled bond as father and daughter.`,
  },

  consensus_rating: 9.2,
  executive_summary: `"The Arrangement" is an exceptional psychological thriller examining artistic exploitation through the story of Sabine Moreau, a young actress who becomes entangled with renowned director Lucien Duret at his 1960s Italian villa.\n\nThe screenplay excels in atmospheric writing, character psychology, and thematic depth. It offers a searing critique of artistic exploitation while maintaining narrative ambiguity about complicity, agency, and the cost of transformation.`,

  strengths: [
    'Exceptional dialogue with sophisticated subtext',
    'Layered characterization - Sabine\'s arc from victim to complicit participant',
    'Sophisticated visual storytelling using mirrors, light, and surveillance',
    'Thematic resonance - explores identity, performance, and the male gaze without didacticism',
    'Period authenticity in 1960s European art cinema setting',
    'Strong opening (screen test) and closing (Cannes premiere) bookends',
    'Career-defining roles for both leads',
  ],

  areas_for_development: [
    'Act Two pacing could be tightened by 8-10 pages (scenes 17, 19, 27, 30)',
    'Some metaphorical dialogue may read as overly literary',
    'Delphine\'s arc deserves one additional scene (between 38-39)',
    'Nico character feels slightly thin - either expand or make more explicitly symbolic',
    'Time passage in villa scenes could be clearer',
  ],

  structural_analysis: {
    page_count: 110,
    act_structure: '3-act',
    act_breakdowns: [
      {
        act_number: 1,
        page_range: '1-35',
        scenes: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14],
        opening_image: 'Screen test - Sabine stripped bare before camera and Lucien\'s voice',
        inciting_incident: 'Invitation to the villa (arrives page 2)',
        turning_point: 'The library kiss (Scene 15, ~page 33)',
        structural_strengths: [
          'Brilliant opening that introduces all themes',
          'Economy of exposition - backstory emerges organically',
          'Visual storytelling dominates over explanation',
          'Environment establishes gothic isolation',
        ],
        pacing_observations: [
          'Opening 20 pages are gripping',
          'Strong world-building through villa descriptions',
          'Character introductions are dynamic, not static',
        ],
      },
      {
        act_number: 2,
        page_range: '35-80',
        scenes: [15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42],
        midpoint: 'Screening room seduction (Scene 20, ~page 45)',
        turning_point: 'Soundstage confrontation with Delphine present (Scene 35, ~page 80)',
        structural_strengths: [
          'Spiral structure - repeating patterns with deepening revelations',
          'Delphine functions as chorus and ghost',
          'Parallel tracking of film production and Sabine\'s psychology',
        ],
        pacing_observations: [
          'Pages 35-60: Strong forward momentum',
          'Pages 60-75: Some repetitiveness in dinner scenes',
          'Pages 75-80: Excellent acceleration to turning point',
        ],
        trim_recommendations: [
          'Scene 17 (Conservatory) - atmospheric but somewhat redundant',
          'Scene 19 (Villa grounds) - establishing shot could be consolidated',
          'Scene 27 (Morning transition) - time passage utility only',
          'Scene 30 (Hallway moment) - brief transition that slows momentum',
        ],
      },
      {
        act_number: 3,
        page_range: '80-110',
        scenes: [43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56],
        climax: 'Filming of breakthrough monologue (Scene 46), villa confrontation (Scene 49)',
        structural_strengths: [
          'Sabine masters Lucien\'s tools and turns them back',
          'Parallel completions - film finished, Sabine transformed',
          'Cannes ending is both triumph and warning',
          'Ambiguous without being frustrating',
        ],
        pacing_observations: [
          'Final 30 pages maintain excellent tension',
          'Resolution is thematically consistent',
          'Leaves room for interpretation without being obtuse',
        ],
      },
    ],
    pacing_rhythm: 'Compression/expansion pattern mirrors Sabine\'s psychology',
    scene_count: 56,
    dialogue_to_action_ratio: '40/60 - Visually driven with literary dialogue',
    structural_innovations: [
      'Non-linear emotional progression',
      'Multiple false peaks where Sabine gains then loses agency',
      'Film-within-film meta-structure',
    ],
    trim_recommendations: [
      {
        section: 'Act Two, Scenes 17-30',
        pages: '8-10 pages',
        rationale: 'Several atmospheric/transitional scenes serve mood but slow commercial pacing',
      },
    ],
  },

  character_analyses: [
    {
      name: 'Sabine Moreau',
      role: 'lead',
      screen_time_percentage: 85,
      arc_description: 'Uncertain ingenue → Observant student → Complicit participant → Self-aware image',
      complexity_notes: 'Sabine avoids simple victimhood. She gains power by mastering the exploitative system rather than escaping it. This moral ambiguity elevates the material.',
      voice_evolution: {
        act_one: [
          'Deferential: "I\'m very grateful to be here"',
          'Questioning: "Is that a flaw?"',
          'Uncertain: "I thought maybe it was a test"',
        ],
        act_two: [
          'Testing boundaries: "That\'s not in the sides"',
          'Self-aware but complicit: "I want to know I\'m not a replacement"',
          'Gaining vocabulary: "Speaking in riddles. Saying I\'m the only one and making me feel interchangeable"',
        ],
        act_three: [
          'Confident: "I needed air" / "Might be good for you to stop watching sometimes"',
          'Articulate about exploitation: "Lucien captures erosion. Not just identity"',
          'Masters his language: Uses metaphor back at him',
        ],
      },
      performance_demands: [
        'Must play both vulnerability and calculation simultaneously',
        'Transformation arc needs subtlety - no melodramatic breakdown',
        'Significant emotional range required',
        'Comfort with intimate scenes',
        'Ability to hold extreme close-ups without "acting"',
      ],
      casting_recommendations: [
        'Saoirse Ronan',
        'Florence Pugh',
        'Anya Taylor-Joy',
        'Thomasin McKenzie',
        'Jessie Buckley',
      ],
      development_opportunities: [
        'Consider one additional scene showing Sabine\'s life before Lucien',
      ],
    },
    {
      name: 'Lucien Duret',
      role: 'lead',
      screen_time_percentage: 70,
      arc_description: 'Charismatic auteur → Possessive controller → Desperate manipulator → Outmaneuvered artist',
      complexity_notes: 'Sophisticated because he genuinely believes his mythology. He\'s not a cartoon villain but an artist who views people as materials. Simultaneously seductive and monstrous.',
      voice_evolution: {
        act_one: [
          'Questions that aren\'t questions: "Do you trust me?"',
          'Observational: "You enter like someone trying not to"',
          'Metaphorical: "Like a piece of music. The notes are already there"',
        ],
        act_two: [
          'Clinical intimacy: "I want to see what happens when you speak without knowing if you\'re alone"',
          'Possessive: References to ownership disguised as artistic direction',
          'Jealous: Barely contained rage over Nico',
        ],
        act_three: [
          'Desperate: "You can\'t leave. The film isn\'t finished"',
          'Manipulative appeals: "Everything I\'ve done has been for your art"',
          'Final acknowledgment: Recognition that she\'s surpassed him',
        ],
      },
      performance_demands: [
        'Must be naturally charismatic - audience must understand the attraction',
        'European sophistication without caricature',
        'Ability to shift from charm to menace subtly',
        'Requires actor who can make intellectual seduction believable',
      ],
      casting_recommendations: [
        'Colin Firth',
        'Jude Law',
        'Oscar Isaac',
        'Michael Fassbender',
        'Ralph Fiennes',
        'Mads Mikkelsen',
      ],
    },
    {
      name: 'Delphine',
      role: 'supporting',
      screen_time_percentage: 15,
      arc_description: 'Resigned prisoner → Cryptic warner → Sabine\'s mirror',
      complexity_notes: 'Understands the game completely but cannot or will not leave. Represents both warning and inevitability. Economical use - appears in few scenes but haunts entire film.',
      voice_evolution: {
        act_one: [
          'Cryptic warnings: "Wait until the wind turns. Like ghosts moaning"',
          'Wounded observations: "You enter like someone who believes they belong"',
        ],
        act_two: [
          'Direct confrontation: "Was the silk tie a new addition?"',
          'Teaching moments: "He gives you the line. You give him your throat"',
        ],
        act_three: [
          'Final warnings go unheeded',
          'Recognition of pattern repeating',
        ],
      },
      performance_demands: [
        'European sophistication and world-weariness',
        'Ability to convey volumes with minimal dialogue',
        'Supporting actress showcase material',
      ],
      casting_recommendations: [
        'Tilda Swinton',
        'Marion Cotillard',
        'Léa Seydoux',
        'Kristin Scott Thomas',
      ],
      development_opportunities: [
        'Add one scene between 38-39 where she explicitly tells Sabine about previous actress who "disappeared"',
        'Would create showcase moment and set up Scene 55 (burned film) with more impact',
      ],
    },
  ],

  thematic_analysis: {
    primary_themes: [
      {
        theme: 'Identity & Performance',
        analysis: 'Central question: Who are we when constantly observed? The screenplay explores how identity becomes performance under the male gaze. Sabine literally doesn\'t know who she is "when no one is watching".',
        sophistication_notes: 'Doesn\'t argue Sabine loses her "true self" (implying one exists) but rather that identity is always performative. The question is who controls the performance.',
      },
      {
        theme: 'Art vs. Exploitation',
        analysis: 'Where is the line between artistic vision and human exploitation? Lucien genuinely believes great art justifies manipulation.',
        sophistication_notes: 'Screenplay refuses to definitively answer whether Sabine\'s transformation produces "great art" or simply documents exploitation. This ambiguity is the point.',
      },
      {
        theme: 'The Male Gaze',
        analysis: 'Both critiques and demonstrates the male gaze without being didactic. Sabine is perpetually observed, filmed, judged, constructed.',
        sophistication_notes: 'Critiques the male gaze BY showing it, not by having characters lecture about it. Visual motifs (mirrors, cameras, white screens) make it literal.',
      },
    ],
    visual_motifs: [
      'Mirrors and reflective surfaces',
      'White walls and screens',
      'Cameras and surveillance',
      'Light quality shifts (warm→cold as Sabine\'s agency decreases)',
      'Sea/nature sounds as emotional weather',
    ],
    symbolic_elements: [
      'The villa as beautiful prison',
      'Costume changes mark identity shifts',
      'Burned film reels = destroyed previous muses',
      'Cannes red carpet = ultimate performance/constructed image',
    ],
    philosophical_position: 'Pragmatic feminist perspective - perfect liberation may be impossible, survival requires mastering oppressive systems rather than escaping them.',
  },

  scene_analyses: [
    {
      scene_number: 1,
      scene_heading: 'INT. SOUNDSTAGE - DAY',
      page_start: 1,
      page_end: 2,
      scene_function: 'Establishes central dynamic: Lucien\'s voice controls the space while Sabine performs vulnerability on command. Inverts traditional screen test - rather than proving she can act, Sabine reveals she may not know who she really is.',
      dialogue_quality_rating: 9,
      dialogue_notes: 'Exceptional. Lucien\'s off-camera dialogue establishes his methodology. Sabine\'s improvised admission is trailer gold.',
      commercial_appeal: {
        festival_circuit: 'Extended uncomfortable screen test creates "hold your breath" tension that plays brilliantly with Cannes/Venice audiences',
        actor_showcase: 'Oscar nomination scene - requires actress who can convey multiple layers simultaneously',
        international: 'French dialogue and European art cinema aesthetic opens doors in France, Italy, UK markets',
      },
      actor_appeal: {
        role_significance: 'Career-defining opening - demonstrates range, vulnerability, and ability to be compelling while essentially motionless',
        comp_casting: ['Jessie Buckley', 'Thomasin McKenzie', 'Anya Taylor-Joy'],
        oscar_potential: 'Oscar clip scene - holds extreme close-ups, navigates power dynamics with subtlety',
      },
      production_considerations: {
        budget_impact: 'LOW',
        shooting_days: '1-2 days maximum',
        location_requirements: 'Single soundstage location',
        vfx_requirements: 'None - period camera equipment only',
      },
      pacing_notes: 'Runs ~2 minutes screen time but feels like eternity (in best way). Establishes contemplative rhythm.',
      marketing_implications: 'Opening hook immediately establishes genre and tone. Final line "That\'s why you\'re here" is haunting.',
      key_exchanges: [
        'LUCIEN: "Say it."\nSABINE: "I don\'t know who I am when no one is watching."\nLUCIEN: "And when someone is?"\nSABINE: "I become... what they want."',
      ],
      strengths: [
        'Immediately establishes entire thematic framework',
        'Character introduction without traditional exposition',
        'Quotable dialogue that appears in trailers',
        '"Water cooler moment" - audiences will debate predator vs artist',
      ],
      concerns: [
        'Pacing risk - some commercial buyers may find "too slow"',
        'Requires audience comfort with ambiguity',
        'Marketing challenge - how to sell without revealing too much',
      ],
    },
    {
      scene_number: 20,
      scene_heading: 'INT. VILLA - PRIVATE SCREENING ROOM - NIGHT',
      page_start: 45,
      page_end: 48,
      scene_function: 'Midpoint turning point - Lucien explicitly names "the arrangement" and Sabine consciously chooses to proceed despite doubts. Transforms her from passive victim to active participant.',
      dialogue_quality_rating: 10,
      dialogue_notes: 'Thematically perfect. Lucien articulates his methodology explicitly: "I want the part of you that you\'ve never used"',
      commercial_appeal: {
        festival_circuit: 'Controversial intimate scene but thematically essential - art house audiences will defend it',
        actor_showcase: 'Requires tremendous trust between actors and director',
        international: 'European sensibility - more acceptable internationally than US market',
      },
      actor_appeal: {
        role_significance: 'Midpoint seduction scene where power dynamics are most explicit',
        comp_casting: ['Chemistry between leads is essential'],
        oscar_potential: 'Controversial but will be discussed - "brave" performance territory',
      },
      production_considerations: {
        budget_impact: 'LOW',
        shooting_days: '2-3 days for intimacy',
        location_requirements: 'Villa screening room - controlled environment',
        vfx_requirements: 'None',
      },
      pacing_notes: 'Extended scene (3 pages) earns intimacy through slowness. Allows breathing room for moral complexity.',
      marketing_implications: 'Won\'t be in trailer but will generate think pieces and controversy in positive way for prestige positioning.',
      key_exchanges: [
        'LUCIEN: "I want the part of you that you\'ve never used."\nSABINE: "What if there isn\'t anything left?"',
      ],
      strengths: [
        'Explicit moment where Sabine becomes complicit',
        'Complicates audience moral relationship with her choices',
        'Thematically consistent with exploitation critique',
      ],
      concerns: [
        'Intimate content may limit commercial distribution',
        'Requires intimacy coordinator and careful handling',
        'Risk of being misread as exploitation rather than critique',
      ],
    },
    {
      scene_number: 56,
      scene_heading: 'EXT. CANNES - PALAIS RED CARPET - EVENING',
      page_start: 108,
      page_end: 110,
      scene_function: 'Perfect thematic resolution - Sabine walks red carpet having become the image Lucien created, but with full awareness. Sees him with new young intern, smiles knowingly, walks away.',
      dialogue_quality_rating: 8,
      dialogue_notes: 'Minimal dialogue - mostly visual storytelling. Final exchange is perfectly ambiguous.',
      commercial_appeal: {
        festival_circuit: 'Meta-textual perfection - film about filmmaking ends at film festival. Cannes will love this.',
        actor_showcase: 'Wordless performance showing transformation complete',
        international: 'Red carpet glamour sells internationally',
      },
      actor_appeal: {
        role_significance: 'Final image defines entire arc - triumph and tragedy simultaneously',
        comp_casting: ['Requires actress who can convey knowing complexity without dialogue'],
        oscar_potential: 'Iconic closing image - will be in Oscar montages',
      },
      production_considerations: {
        budget_impact: 'MEDIUM',
        shooting_days: '3-4 days for red carpet recreation',
        location_requirements: 'Cannes Palais or convincing double - may require location shoot',
        vfx_requirements: 'Crowd replication, period flashbulbs',
      },
      pacing_notes: 'Slowed-down glamour contrasts with villa claustrophobia. Visual catharsis.',
      marketing_implications: 'This IS the poster image - Sabine in couture on red carpet, Lucien blurred in background.',
      key_exchanges: [
        '[Sabine sees Lucien with young intern]\nShe meets his eyes. Smiles. A little too long. A little too knowingly.\nThen turns and walks into the theater. He starts to follow—\nThe doors close. She\'s gone.',
      ],
      strengths: [
        'Visually iconic closing image',
        'Ambiguous without being frustrating',
        'Cycle continuing suggests pattern is systemic, not individual',
        'Sabine\'s smile shows she understands and chooses to move forward anyway',
      ],
      concerns: [
        'Budget required for period Cannes recreation',
        'Requires strong visual effects for authentic period feel',
      ],
    },
  ],

  industry_intelligence: {
    comp_titles: [
      {
        title: 'Phantom Thread',
        year: 2017,
        similarity_percentage: 95,
        comparison_notes: 'Power dynamics, European elegance, psychological games, obsessive artist relationship',
      },
      {
        title: 'The Master',
        year: 2012,
        similarity_percentage: 88,
        comparison_notes: 'Psychological manipulation, mentor-protégé power imbalance, ambiguous morality',
      },
      {
        title: 'Contempt',
        year: 1963,
        similarity_percentage: 85,
        comparison_notes: 'Setting/era/meta-cinema, Italian villa, marriage of convenience, film industry critique',
      },
      {
        title: '8½',
        year: 1963,
        similarity_percentage: 82,
        comparison_notes: '1960s Italian film culture, auteur as god-figure, women as muses',
      },
      {
        title: 'Black Swan',
        year: 2010,
        similarity_percentage: 78,
        comparison_notes: 'Artistic obsession, psychological breakdown, identity dissolution, transformation costs',
      },
    ],
    market_position: 'Prestige arthouse with strong commercial potential - sits at intersection of European Art Cinema, Psychological Power Dramas, and Gothic Romance',
    budget_range: '$8M-$15M',
    revenue_projection: '$25M-$65M worldwide',
    awards_potential: 'VERY HIGH',
    festival_strategy: 'Cannes premiere ESSENTIAL - Competition or Un Certain Regard. Venice and Telluride as alternatives.',
    target_distributors: [
      'A24',
      'Neon',
      'Focus Features',
      'Fox Searchlight',
      'Sony Pictures Classics',
    ],
    casting_tier: 'A-list leads essential for financing and awards positioning',
  },

  dialogue_analysis: {
    overall_quality: 'Exceptional - literary without being theatrical, realistic while maintaining heightened poetry appropriate to period and genre.',
    character_voices: [
      {
        character: 'Lucien',
        voice_description: 'Master manipulator who uses questions as statements, metaphor as deflection, and clinical language for intimate moments',
        examples: [
          '"Do you trust me?" (not actually a question)',
          '"You enter like someone trying not to" (observation as judgment)',
          '"Like a piece of music. The notes are already there. You just have to learn how to be played"',
        ],
      },
      {
        character: 'Sabine',
        voice_description: 'Evolves from deferential and uncertain to confident and articulate, eventually mastering Lucien\'s metaphorical language',
        examples: [
          'Early: "I\'m very grateful to be here"',
          'Middle: "Speaking in riddles. Saying I\'m the only one and making me feel interchangeable"',
          'Late: "Lucien captures erosion. Not just identity"',
        ],
      },
    ],
    subtext_examples: [
      {
        scene: 4,
        exchange: 'DELPHINE: "You look like someone who sings when no one\'s watching."',
        surface: 'Small talk about Sabine\'s appearance',
        subtext: 'You still have a private self. You won\'t for long.',
      },
      {
        scene: 15,
        exchange: 'LUCIEN: "What would you give to make her real?"\nSABINE: "I\'m already giving all of it."',
        surface: 'Discussion of the role',
        subtext: 'She\'s already surrendering her identity; he\'s measuring her awareness of this',
      },
      {
        scene: 56,
        exchange: '[Sabine sees Lucien with young intern. She meets his eyes. Smiles. Too long. Too knowingly.]',
        surface: 'Acknowledgment',
        subtext: 'I see you. I know the pattern. I\'m complicit now. And I\'m walking away anyway.',
      },
    ],
  },

  visual_storytelling: {
    overall_assessment: 'Exceptionally visual - uses action lines and environmental description to convey psychology. Written by someone who understands cinema.',
    techniques: [
      'Character revealed through physical behavior (Sabine\'s trembling hands, Delphine\'s smoking, Lucien\'s stillness)',
      'Environment as emotion (light quality, architectural spaces mirror psychological spaces)',
      'Camera-aware writing with specific shot descriptions',
    ],
    environment_as_emotion: [
      '"Blindingly beautiful, almost unreal" - early hope and seduction',
      '"Low lamplight casts a warm haze" - false intimacy',
      '"The light is colder now" - psychological shift',
      'Open courtyards for early hope, narrow hallways for paranoia',
    ],
    camera_aware_notes: [
      '"The camera lingers. Too long. Sabine\'s composure wavers."',
      '"We hear the rustling of reel tape, quiet French murmurs off-camera."',
      'Specific shot compositions described for key moments',
    ],
  },

  author_profile: {
    estimated_age_range: 'Late 20s to late 30s',
    education_indicators: 'Graduate-level film/literature education - demonstrates sophisticated understanding of European cinema history',
    sophistication_level: 'Professional-ready, expert-level craft - NOT a beginner',
    philosophical_position: 'Pragmatic feminist - understands perfect liberation may be impossible within patriarchal systems',
    industry_readiness: 'Ready for major representation (WME, CAA, UTA level)',
  },

  recommendation: {
    verdict: 'RECOMMEND',
    summary: `"The Arrangement" is exceptional psychological thriller that merits strong representation and aggressive festival/awards positioning. The screenplay demonstrates mastery of character psychology, sophisticated visual storytelling, and thematic depth that elevates it above typical genre fare.\n\nWith proper casting (A-list leads essential) and strategic festival launch (Cannes premiere ideal), this has the potential for significant commercial success within the prestige space and multiple Oscar nominations.\n\nCompete directly with films like "Phantom Thread" and "The Master" for audience and awards attention.`,
    next_steps: [
      'Secure A-list representation immediately',
      'Target prestige financiers (Plan B, Scott Rudin, Working Title)',
      'Strategic casting begins with Sabine role - this is career-defining',
      'Festival strategy: Cannes Competition submission',
      'Minor script polish: trim 8-10 pages from Act Two, add Delphine scene',
      'Author should be involved in adaptation process',
    ],
  },

  created_at: '2025-11-20',
  updated_at: '2025-11-20',
  version: 1,
  status: 'completed',
}
