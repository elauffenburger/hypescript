use std::{
    env::{self},
    fmt::Display,
    fs,
    io::{self, BufRead, Read},
};

use hypescript::{emitter, parser};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = parse_args()?;
    let output_dir = args.output_dir.expect("output dir required!");
    let mut src_reader: Box<dyn BufRead> = match args.source {
        Some(source) => match source {
            Source::Stdin => Box::new(io::BufReader::new(io::stdin().lock())),
            Source::Path(path) => Box::new(io::BufReader::new(fs::File::open(path)?)),
        },
        None => Box::new(io::BufReader::new(io::stdin().lock())),
    };

    let mut src = String::new();
    src_reader.read_to_string(&mut src)?;

    let parser = parser::Parser::new(".".into());
    let emitted = emitter::Emitter::new().emit(&[parser.parse(&src)?])?;

    for file in &emitted.files {
        write_emitted_file(&file, &output_dir)?;
    }

    Ok(())
}

fn write_emitted_file(
    file: &emitter::EmittedFile,
    dir: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    Ok(match file {
        emitter::EmittedFile::Dir { name, files } => {
            let dir = &format!("{dir}/{name}");
            fs::create_dir_all(dir)?;

            for file in files {
                write_emitted_file(&file, dir)?;
            }
        }
        emitter::EmittedFile::File { name, content } => {
            fs::create_dir_all(dir)?;

            fs::write(format!("{dir}/{name}"), content)?;
        }
    })
}

struct ParseArgsError {
    msg: String,
}

impl std::fmt::Debug for ParseArgsError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(&self.msg)
    }
}

impl Display for ParseArgsError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(&self.msg)
    }
}

impl std::error::Error for ParseArgsError {}

impl ParseArgsError {
    pub fn new(msg: &str) -> Box<ParseArgsError> {
        Box::new(ParseArgsError { msg: msg.into() })
    }
}

enum Source {
    Stdin,
    Path(String),
}

struct HscArgs {
    output_dir: Option<String>,
    source: Option<Source>,
}

fn parse_args() -> Result<HscArgs, Box<dyn std::error::Error>> {
    let mut args = HscArgs {
        output_dir: None,
        source: None,
    };

    let mut env_args = env::args().skip(1);
    while let Some(arg) = env_args.next() {
        match arg.as_str() {
            "-o" => match env_args.next() {
                Some(output_dir) => {
                    args.output_dir = Some(output_dir);
                }
                None => return Err(ParseArgsError::new("arg for -o required")),
            },
            _ => match args.source {
                None => {
                    args.source = Some(if arg == "-" {
                        Source::Stdin
                    } else {
                        Source::Path(arg)
                    })
                }
                Some(_) => return Err(ParseArgsError::new(&format!("unknown arg {arg}"))),
            },
        }
    }

    Ok(args)
}
