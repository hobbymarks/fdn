use std::{
    fs::{self, File},
    hint::black_box,
    io::{Result, Write},
    path::{Path, PathBuf},
};

use criterion::{criterion_group, criterion_main, Criterion};
use fdn::{directories, regular_files};
use tempfile::TempDir;

/// The temporary directory and its contents will be automatically
/// deleted when the returned `TempDir` object goes out of scope.
fn create_temp_files_and_dir() -> Result<(PathBuf, TempDir)> {
    // Create a new temporary directory
    let tmp_dir = TempDir::new()?;
    let root_path = tmp_dir.path().to_owned();

    // Create a subdirectory inside the temporary directory
    let sub_dir_path = root_path.join("my_subdir");
    fs::create_dir(&sub_dir_path)?;

    // Create a file in the root temporary directory
    let f1st_path = root_path.join("file1.txt");
    let mut f1st = File::create(&f1st_path)?;
    writeln!(f1st, "Content of file 1")?;

    // Create a file in the subdirectory
    let f2nd_path = sub_dir_path.join("file2.txt");
    let mut f2nd = File::create(&f2nd_path)?;
    writeln!(f2nd, "Content of file 2")?;

    Ok((root_path, tmp_dir))
}

fn fdn_benchmark_function(c: &mut Criterion) {
    let mut group = c.benchmark_group("Walk dir");

    group.bench_function("fdn_regular_files", |b| match create_temp_files_and_dir() {
        Ok((path, _temp_hanle)) => b.iter(|| black_box(return_regular_files(Some(path.clone())))),
        Err(_) => todo!(),
    });

    group.bench_function("fdn_directories", |b| match create_temp_files_and_dir() {
        Ok((path, _temp_hanle)) => b.iter(|| black_box(return_directories(Some(path.clone())))),
        Err(_) => todo!(),
    });
}

fn return_regular_files(path: Option<PathBuf>) {
    match path {
        Some(p) => {
            let _ = regular_files(&p, 1, vec![]);
        }
        None => {
            let path = Path::new(".");
            let _ = regular_files(path, 1, vec![]);
        }
    }
}

fn return_directories(path: Option<PathBuf>) {
    match path {
        Some(p) => {
            let _ = directories(&p, 1, vec![]);
        }
        None => {
            let path = Path::new(".");
            let _ = directories(path, 1, vec![]);
        }
    }
}

criterion_group!(benches, fdn_benchmark_function);
criterion_main!(benches);
