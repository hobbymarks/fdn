use std::{
    error::Error,
    path::{Path, PathBuf},
};
use std::{fs, process};

use clap::Parser;

/// Simple CLI that accepts a single input path (file or directory).
#[derive(Debug, Parser)]
#[command(
    name = "path-checker",
    version,
    about = "Detects whether the provided input path is a file or a directory.",
    author
)]
struct Args {
    /// Input path to a file or directory
    input: PathBuf,
}

fn walk_dir(input: &Path) -> Result<Vec<PathBuf>, Box<dyn Error>> {
    // Collect absolute paths of all subdirectories and all regular files under `input`.
    let mut results: Vec<PathBuf> = Vec::new();

    // Start from the canonical (absolute) path of the input
    let base = input.canonicalize()?;

    // We want to traverse inside the directory. If input isn't a directory, still try to
    // treat it as a single path entry.
    if base.is_dir() {
        for entry in fs::read_dir(&base)? {
            let path = entry?.path();
            let meta = fs::symlink_metadata(&path)?;
            if meta.is_dir() {
                results.push(path.canonicalize()?);
                // Recurse into subdirectories
                let mut sub = walk_dir(&path)?;
                results.append(&mut sub);
            } else if meta.is_file() {
                results.push(path.canonicalize()?);
            }
            // Other types (symlinks, etc.) are ignored unless they are regular files/directories.
        }
    } else {
        // If it's a file, return its canonical path
        results.push(base);
    }

    Ok(results)
}

fn main() {
    let args = Args::parse();

    // Use symlink_metadata so symlinks are recognized and not automatically followed.
    let metadata = match fs::symlink_metadata(&args.input) {
        Ok(m) => m,
        Err(err) => {
            eprintln!("Error: cannot access '{}': {}", args.input.display(), err);
            process::exit(2);
        }
    };

    // Demonstration: walk the directory and collect all subdirectories and files.
    match walk_dir(&args.input) {
        Ok(paths) => {
            println!("Found {} path(s):", paths.len());
            for p in paths {
                println!("{}", p.display());
            }
        }
        Err(err) => {
            eprintln!("Error while walking directory: {}", err);
            process::exit(3);
        }
    }

    if metadata.is_dir() {
        println!("'{}' is a directory.", args.input.display());
    } else if metadata.is_file() {
        println!("'{}' is a file.", args.input.display());
    } else {
        println!(
            "'{}' exists but is neither a regular file nor a directory.",
            args.input.display()
        );
    }
}
